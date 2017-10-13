package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/go-chi/chi"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/kelseyhightower/envconfig"
)

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

type Config struct {
	Prefix      string
	ServiceName string `envconfig:"NAME" default:"sampleservice"`
	ServiceID   string `envconfig:"ID" required:"true"`
	ServicePort int    `envconfig:"PORT" default:"3344"`
}

// Load settles ENV variables into Config structure
func (c *Config) Load(prefix string) error {
	return envconfig.Process(prefix, c)
}

type Service struct {
	cfg              *Config
	httpServer       *http.Server
	mux              *chi.Mux
	maintenanceMode  bool
	ctx              context.Context
	consulClient     *consulapi.Client
	consulServiceReg *consulapi.AgentServiceRegistration
	serviceID        string
	sync.RWMutex
}

func NewService(ctx context.Context, cfg *Config) *Service {

	consulClient, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		panic(err)
	}

	serviceRegistration := &consulapi.AgentServiceRegistration{
		ID:      cfg.ServiceName + cfg.ServiceID,
		Name:    cfg.ServiceName,
		Port:    cfg.ServicePort,
		Address: GetOutboundIP().String(),
	}

	mux := chi.NewRouter()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serviceName := os.Getenv("SAMPLE_SERVICE_NAME") + os.Getenv("SAMPLE_SERVICE_ID")
		w.Write([]byte("I am " + serviceName + "\n"))
	})

	return &Service{
		ctx:              ctx,
		cfg:              cfg,
		mux:              mux,
		serviceID:        cfg.ServiceName + cfg.ServiceID,
		consulServiceReg: serviceRegistration,
		consulClient:     consulClient,
	}
}

func (r *Service) listen(enableMaintenance bool) error {

	var m *chi.Mux

	if !enableMaintenance {
		m = r.mux
	} else {
		m = chi.NewRouter()
		maintenancePage := func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Service is down for maintenance. Sorry!"))
		}
		m.Get("/", maintenancePage)
		m.NotFound(maintenancePage)
	}

	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(r.cfg.ServicePort),
		Handler: m,
	}

	r.httpServer = httpServer

	log.Printf("Starting %v Server", r.cfg.ServiceName)

	return r.httpServer.ListenAndServe()
}

// Listen starts the http server to listen on the designated port
// load configuration here to support reload
func (r *Service) Listen() error {
	if err := r.Register(); err != nil {
		return err
	}
	return r.listen(false)
}

func (r *Service) Register() error {
	return r.consulClient.Agent().ServiceRegister(r.consulServiceReg)
}

func (r *Service) DeRegister() error {
	return r.consulClient.Agent().ServiceDeregister(r.consulServiceReg.ID)
}

func (r *Service) Reload() error {
	r.Lock()
	defer r.Unlock()

	// shutdown http server
	err := r.httpServer.Shutdown(r.ctx)
	if err != nil {
		return err
	}

	// reload config
	cfg := &Config{}
	if r.cfg.Prefix == "" {
		log.Fatal(fmt.Errorf("Config not initialized with environment var prefix"))
	}
	if err := cfg.Load(r.cfg.Prefix); err != nil {
		log.Fatal(err)
	}

	r.cfg = cfg

	return r.listen(false)
}

func (r *Service) Maintenance() error {
	r.Lock()
	defer r.Unlock()

	r.maintenanceMode = !r.maintenanceMode

	// shutdown http server
	err := r.httpServer.Shutdown(r.ctx)
	if err != nil {
		return err
	}
	return r.listen(r.maintenanceMode)
}

// Shutdown stops the http server gracefully
func (r *Service) Shutdown() error {
	r.Lock()
	defer r.Unlock()
	if err := r.DeRegister(); err != nil {
		fmt.Println(err)
	}
	return r.httpServer.Shutdown(r.ctx)
}
