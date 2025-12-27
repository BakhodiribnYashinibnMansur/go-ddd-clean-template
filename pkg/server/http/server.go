package http

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/mbndr/figlet4go"
)

// Server is a struct that represents the server
type Server struct {
	httpServer *http.Server
}

func NewServer() *Server {
	return &Server{}
}

// Run is a function that runs the server
func (s *Server) Run(port int, handler http.Handler) error {
	for {
		status, err := checkPortBind(port)
		if err != nil {
			port++
		}
		if status {
			break
		}
	}
	figlet(port)
	s.httpServer = &http.Server{
		Addr:           ":" + strconv.Itoa(port),
		Handler:        handler,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
	}

	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Shutdown is a function that shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Check if a port is available
func checkPortBind(port int) (status bool, err error) {
	// Concatenate a colon and the port
	host := ":" + strconv.Itoa(port)
	// Try to create a server with the port
	server, err := net.Listen("tcp", host)
	// if it fails then the port is likely taken
	if err != nil {
		return false, err
	}
	// close the server
	server.Close()
	// we successfully used and closed the port
	// so it's now available to be used again
	return true, nil
}

func figlet(port int) {
	ascii := figlet4go.NewAsciiRender()
	// Adding the colors to RenderOptions
	options := figlet4go.NewRenderOptions()
	options.FontName = "larry3d"
	options.FontColor = []figlet4go.Color{
		// Colors can be given by default ansi color codes...
		figlet4go.ColorGreen,
		// figlet4go.ColorYellow,
		figlet4go.ColorCyan,
		// figlet4go.ColorMagenta,
		// figlet4go.ColorWhite,
		figlet4go.ColorRed,
		figlet4go.ColorBlue,
		figlet4go.ColorBlack,
		// ...or by an rgb value
		// figlet4go.Color{R: 255, G: 0, B: 0},
		// ...or by an hex string...
		// figlet4go.NewTrueColorFromHexString("885DBA"),
		// ...or by an TrueColor object with rgb values
		// figlet4go.TrueColor{136, 93, 186},
	}
	renderStr, err := ascii.RenderOpts(strconv.Itoa(port), options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server is running on port :")
	fmt.Print(renderStr)
}
