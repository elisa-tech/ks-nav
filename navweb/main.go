package main

import (
	"bufio"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"net/http"
	"os/exec"
	"html/template"
	"github.com/gorilla/websocket"
	"fmt"
)

type ImageGenerationRequest struct {
	StartSymbol string `form:"startsymbol" binding:"required"`
	Instance    int    `form:"instance" binding:"required"`
	DisplayMode string `form:"displaymode"`
	Depth       string `form:"depth"`
}
var upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
                return true
        },
}

func handleWebSocket(c *gin.Context) {
	log.Println("handleWebSocket - start")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	log.Println("handleWebSocket - execute")
	cmd := exec.Command("nav-db-filler")
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Error creating stdout pipe:", err)
		return
	}

	go func() {
		log.Println("handleWebSocket - fetcher")
		scanner := bufio.NewScanner(cmdReader)
		for scanner.Scan() {
			line := scanner.Text()
			if err := conn.WriteMessage(websocket.TextMessage, []byte(line+"\n")); err != nil {
				log.Println("Error sending WebSocket message:", err)
				return
			}
		}
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Execution Terminated!\n")); err != nil {
			log.Println("Error sending WebSocket message:", err)
			return
		}
	}()

	log.Println("handleWebSocket - check")
	if err := cmd.Start(); err != nil {
		log.Println("Error starting command:", err)
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Println("Error waiting for command:", err)
		return
	}
}

func loadHTMLFromBytes(router *gin.Engine, templateName string, templateString []byte) {
	fmt.Println("registering: ", templateName)
	tmpl := template.New(templateName)
	tmpl, err := tmpl.Parse(string(templateString))
	if err != nil {
		panic(err)
	}
	router.SetHTMLTemplate(tmpl)
}

func generateImageHandler(c *gin.Context) {
	var request ImageGenerationRequest

	if err := c.ShouldBind(&request); err != nil {
		log.Println("Error binding form data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	cmdstr:=fmt.Sprintf("nav -f /tmp/container.json -s %s -i %d -m %s -x %s -g 1 -j graphOnly |dot -T svg", request.StartSymbol, request.Instance, request.DisplayMode, request.Depth)
	cmd := exec.Command("/bin/sh", "-c", cmdstr)
	fmt.Println("Executing command:", cmd.String())
	output, err := cmd.Output()
	if err != nil {
		log.Println("Error executing command:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate image"})
		return
	}

	c.Header("Content-Type", "image/svg+xml")
	c.String(200, string(output))
}

func main() {
	router := gin.Default()

	fetchws, err := Asset("data/templates/fetch-ws.html")
	if err != nil {
		panic(err)
	}

	explore, err := Asset("data/templates/explore.html")
	if err != nil {
		panic(err)
	}

	index, err := Asset("data/templates/index.html")
	if err != nil {
		panic(err)
	}

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
	router.GET("/explore", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", explore)
	})
	router.GET("/acquire", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", fetchws)
	})

	router.GET("/ws", handleWebSocket)

	router.POST("/generate-image", generateImageHandler)

	conf_file, err := Asset("data/configs/container.json")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("/tmp/container.json", conf_file, 0644)

	router.Run(":8080")
}

