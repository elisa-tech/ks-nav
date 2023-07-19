package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os/exec"
//	"strconv"
	"fmt"
)

type ImageGenerationRequest struct {
	StartSymbol string `form:"startsymbol" binding:"required"`
	Instance    int    `form:"instance" binding:"required"`
	DisplayMode string `form:"displaymode"`
	Depth       string `form:"depth"`
}

func main() {
	router := gin.Default()

	router.LoadHTMLFiles("templates/index.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.POST("/generate-image", generateImageHandler)

	router.Run(":8080")
}

func generateImageHandler(c *gin.Context) {
	var request ImageGenerationRequest

	if err := c.ShouldBind(&request); err != nil {
		log.Println("Error binding form data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

//	cmd := exec.Command("./nav/nav",  "-f", "confs/conf_sample.json", "-s", request.StartSymbol, "-i", strconv.Itoa(request.Instance), "-m", request.DisplayMode, "-x", request.Depth, "-g", "4", "-j","graphOnly")
	cmdstr:=fmt.Sprintf("./nav/nav -f confs/conf_sample.json -s %s -i %d -m %s -x %s -g 1 -j graphOnly |dot -Tsvg", request.StartSymbol, request.Instance, request.DisplayMode, request.Depth)
	cmd := exec.Command("/bin/bash", "-c", cmdstr)
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
