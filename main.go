package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"
)

const (
	BUFLEN   = 4096 * 4
	TIMEOUT  = 60
	RESPONSE = "HTTP/1.1 101 Switching Protocol\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: foo\r\n\r\n"
)

// Config adalah struktur untuk menyimpan konfigurasi dari file YAML
type Config struct {
	Address   string `yaml:"address"`
	PortWS    int    `yaml:"portws"`
	PortHTTPS int    `yaml:"porthttps"`
	Cert      string `yaml:"cert"`
	Key       string `yaml:"key"`
}

// ServerHandler represents the proxy server
type ServerHandler struct {
	Config Config
	upgrader websocket.Upgrader
	logger   *log.Logger
}

// NewServerHandler initializes a new instance of ServerHandler
func NewServerHandler(config Config) *ServerHandler {
	return &ServerHandler{
		Config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		logger: getLogger(),
	}
}

// getLogger creates a logger with file at /var/log/goproxy.log
func getLogger() *log.Logger {
	// Open or create the log file
	file, err := os.OpenFile("/var/log/goproxy.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	// Create a logger with file as the output
	logger := log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	return logger
}

// Run starts the proxy server
func (s *ServerHandler) Run() {
	http.HandleFunc("/", s.handleWebSocket)
	listenAddr := fmt.Sprintf("%s:%d", s.Config.Address, s.Config.PortWS)
	s.logger.Printf("Proxy server is listening on %s\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func (s *ServerHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Printf("Error upgrading to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	for {
		// Baca pesan dari klien
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.logger.Printf("Error reading message: %v", err)
			break
		}

		// Ubah pesan menjadi huruf besar
		resp := strings.ToUpper(string(msg))

		// Kirim balasan ke klien
		err = conn.WriteMessage(websocket.TextMessage, []byte(resp))
		if err != nil {
			s.logger.Printf("Error writing message: %v", err)
			break
		}
	}
}

// Usage menampilkan cara penggunaan aplikasi
func Usage() {
	fmt.Println("Websocket Goproxy By FN Project")
	fmt.Println("Report Bug Mail: lumine@rerechan02.com")
	fmt.Println("\nUsage:")
	flag.PrintDefaults()
}

func main() {
	var configPath string

	// Parse command line flags
	flag.StringVar(&configPath, "config", "", "Path to configuration file (yaml)")
	flag.StringVar(&configPath, "c", "", "Path to configuration file (yaml)")
	help := flag.Bool("h", false, "Display usage instructions")
	flag.BoolVar(help, "help", false, "Display usage instructions")
	flag.Parse()

	// If -h or --help is provided, display usage instructions
	if *help {
		Usage()
		return
	}

	// If -c or --config is not provided, display error and usage instructions
	if configPath == "" {
		fmt.Println("Error: Please provide path to configuration file using -c or --config flag.")
		Usage()
		return
	}

	// Read configuration file
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	// Parse configuration file
	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Error parsing configuration file: %v", err)
	}

	// Display parsed configuration
	fmt.Println("Parsed Configuration:")
	fmt.Println("Address:", config.Address)
	fmt.Println("Port WS:", config.PortWS)
	fmt.Println("Port HTTPS:", config.PortHTTPS)
	fmt.Println("Cert:", config.Cert)
	fmt.Println("Key:", config.Key)

	// Initialize and run the server
	server := NewServerHandler(config)
	server.Run()
}
