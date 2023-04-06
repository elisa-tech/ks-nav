#DEBUG=-g

all: dotv
dot.tab.h: dot.y
	bison -t -v -d -g --report=all -p "dot" dot.y
dot.tab.c: dot.y
	bison -t -v -d -g --report=all -p "dot" dot.y
lex.dot.c: dot.l
	flex -Pdot dot.l
dotv: lex.dot.c dot.tab.c dot.tab.h dotv.c
	gcc $(DEBUG) -o dotv lex.dot.c dot.tab.c dotv.c

libdotparser.so: lex.dot.c dot.tab.c dot.tab.h
	gcc  $(DEBUG) -shared -fPIC -o libdotparser.so lex.dot.c dot.tab.c
clean:
	rm -f dot.output dot.tab.c dot.tab.h dot.tab.o lex.dot.o dot.dot lex.dot.c dotv libdotparser.so
