package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type Mirror struct {
	Name              string          `json:"name"`
	Protocol          string          `json:"protocol"`
	Client            bool            `json:"client"`
	Addr1             string          `json:"addr1"`
	Addr2             string          `json:"addr2"`
	InOut1            string          `json:"in_out1"`
	InOut2            string          `json:"in_out2"`
	Conn1             *SecureConn     `json:"-"`
	Conn2             *SecureConn     `json:"-"`
	ChannelNewConn1   chan SecureConn `json:"-"`
	ChannelNewConn2   chan SecureConn `json:"-"`
	channelNewConnCmd chan SecureConn `json:"-"`
	ChannelEventExit  chan bool       `json:"-"`
	channelEventExit1 chan bool       `json:"-"`
	channelEventExit2 chan bool       `json:"-"`
}

// https://gist.github.com/sevkin/96bdae9274465b2d09191384f86ef39d
func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func ConnectTo(protocol, addr string) (*SecureConn, error) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Println("(ConnectTo)", err.Error())
		return nil, err
	}
	sconn := NewSecureConn(conn)
	if protocol == "tcp" {
		err = sconn.SetNoDelay(true)
		if err != nil {
			log.Println("(ConnectTo) SetNoDelay", err.Error())
			return nil, err
		}
		err = sconn.SetKeepAlive(true)
		if err != nil {
			log.Println("(ConnectTo) SetKeepAlive", err.Error())
			return nil, err
		}
		err = sconn.SetKeepAlivePeriod(3 * time.Second)
		if err != nil {
			log.Println("(ConnectTo) SetKeepAlivePeriod", err.Error())
			return nil, err
		}
		err = sconn.SetLinger(0)
		if err != nil {
			log.Println("(ConnectTo) SetLinger", err.Error())
			return nil, err
		}
		err = sconn.SetReadBuffer(16384)
		if err != nil {
			log.Println("(ConnectTo) SetReadBuffer", err.Error())
			return nil, err
		}
		err = sconn.SetWriteBuffer(16384)
		if err != nil {
			log.Println("(ConnectTo) SetWriteBuffer", err.Error())
			return nil, err
		}
	}
	return sconn, nil
}

func AcceptConn(network, addr string, channelNewConn chan SecureConn, channelEventExit chan bool) {
	listener, err := net.Listen(network, addr)
	if err != nil {
		log.Println(err)
	}
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("(AcceptConn)", err.Error())
				channelEventExit <- true
				break
			}
			defer conn.Close()
			channelNewConn <- *NewSecureConn(conn)
		}
	}()
	<-channelEventExit
	fmt.Println("(AcceptConn) Exit")
}

func (m *Mirror) Start() {
	go m.Handler()

	m.ChannelNewConn1 = make(chan SecureConn)
	m.ChannelNewConn2 = make(chan SecureConn)
	m.channelNewConnCmd = make(chan SecureConn)
	m.ChannelEventExit = make(chan bool, 1)
	m.channelEventExit1 = make(chan bool, 1)
	m.channelEventExit2 = make(chan bool, 1)

	m.InOut1 = "out"
	m.InOut2 = "out"
	_, err := strconv.Atoi(m.Addr1)
	if err == nil {
		m.InOut1 = "in"
	}
	_, err = strconv.Atoi(m.Addr2)
	if err == nil {
		m.InOut2 = "in"
	}

	fmt.Print(m.InOut1)

	if m.InOut1 == "in" {
		go AcceptConn(m.Protocol, ":"+m.Addr1, m.ChannelNewConn1, m.channelEventExit1)
	}

	if m.InOut2 == "in" {
		go AcceptConn(m.Protocol, ":"+m.Addr2, m.ChannelNewConn2, m.channelEventExit2)
	}

	if m.InOut1 == "out" {
		log.Println(m.Name, "(Start) CMD ConnectTo", m.Addr1)
		sconn, err := ConnectTo(m.Protocol, m.Addr1)
		if err != nil {
			log.Println(m.Name, "(Handler Conn1)", err.Error())
			return
		}
		m.Conn1 = sconn
		m.channelNewConnCmd <- *m.Conn1

		if m.Client {
			log.Println(m.Name, "(HandlerCmd) Client enviando AUTH_TOKEN")
			_, e := sconn.WriteFrame(SecureFrame{Type: MsgTypeAuth, Payload: []byte(os.Getenv("AUTH_TOKEN"))})
			if e != nil {
				log.Println(m.Name, "(HandlerCmd) Error enviando AUTH_TOKEN:", e)
				return
			}
		}

	}

}

func (m *Mirror) Handler() {
	log.Println("(Handler) Start")
	for {
		select {
		case connCmd := <-m.channelNewConnCmd:
			log.Println(m.Name, "(Handler) New ConnCMD")
			go m.HandlerCmd(&connCmd)
		case sconn1 := <-m.ChannelNewConn1:
			log.Println(m.Name, "(Handler) New Conn1")
			go m.Handler1(&sconn1)

		case conn2 := <-m.ChannelNewConn2:
			m.Conn2 = &conn2
			log.Println(m.Name, "(Handler) New Conn2")
			m.Conn2.Read(nil)

		case <-m.ChannelEventExit:
			fmt.Println("(Handler) Exit")
			if m.Conn1 != nil {
				(*m.Conn1).Close()
			}
			if m.Conn2 != nil {
				(*m.Conn2).Close()
			}
			go func() {
				m.channelEventExit1 <- true
			}()
			go func() {
				m.channelEventExit2 <- true
			}()
			return
		}
	}
}

// Dentro de HandlerCmd añadir trazabilidad
func (m *Mirror) HandlerCmd(conn1 *SecureConn) {
	log.Println(m.Name, "(HandlerCmd) Start - Client:", m.Client)

	for {
		log.Printf("%s (HandlerCmd) Esperando mensaje...", m.Name)
		msg, err := ReceiveControlMessage(conn1)
		if err != nil {
			log.Println(m.Name, "(HandlerCmd) Error decodificando control message:", err)
			return
		}
		ip := conn1.RemoteAddr().(*net.TCPAddr).IP.String()
		port := strconv.Itoa(msg.Port)

		fmt.Println(m.Name, "(HandlerCmd) Read", ip+":"+port)

		conn, err := net.Dial(m.Protocol, ip+":"+port)
		if err != nil {
			fmt.Println(m.Name, "(HandlerCmd) Dial", err.Error())
			return
		}
		sconn := NewSecureConn(conn)
		m.ChannelNewConn1 <- *sconn
	}
}

func (m *Mirror) Handler1(sconn1 *SecureConn) {
	fmt.Println(m.Name, "(Handler1) Start -", m.Name)
	var sconn2 *SecureConn

	// Si ya tenemos Conn2
	if m.Conn2 != nil {
		sconn2 = m.Conn2
	}

	// OUT: nos conectamos directamente
	if m.InOut2 == "out" {
		rawConn, err := net.Dial(m.Protocol, m.Addr2)
		if err != nil {
			log.Println(m.Name, "(Handler1) Dial error:", err)
			return
		}
		sconn2 = NewSecureConn(rawConn)

	} else {
		// IN: esperamos conexión entrante luego de enviar control message
		port, _ := GetFreePort()
		log.Printf("%s (Handler1) puerto dinámico: %d", m.Name, port)

		// Enviamos mensaje de control cifrado
		if err := SendControlMessage(m.Conn2, port); err != nil {
			log.Println(m.Name, "(Handler1) Error al enviar control message:", err)
			return
		}
		log.Printf("%s (Handler1) Mensaje enviado a Conn2", m.Name)

		listener, err := net.Listen(m.Protocol, fmt.Sprintf(":%d", port))
		if err != nil {
			log.Println(m.Name, "(Handler1) Error escuchando:", err)
			return
		}
		defer listener.Close()

		conn2, err := listener.Accept()
		if err != nil {
			log.Println(m.Name, "(Handler1) Error aceptando:", err)
			return
		}
		sconn2 = NewSecureConn(conn2) // ✅ usa constructor con buffer inicial
		//sconn2 = NewSecureConn(conn2)
	}

	// Validaciones defensivas
	if sconn1 == nil || sconn2 == nil {
		log.Println(m.Name, "(Handler1) Una de las conexiones es nil")
		return
	}
	log.Printf("%s (Handler1) sconn1: %T, sconn2: %T", m.Name, sconn1, sconn2)

	// Canal de sincronización para cerrar cuando termine
	exit := make(chan bool, 2)

	go func() {
		log.Println(m.Name, "(Handler1) Iniciando pipe → A", m.Client)
		if m.Client {
			if _, err := io.Copy(sconn1, sconn2.Conn); err != nil {
				log.Println(m.Name, "pipe A error:", err)
			}
		} else {
			if _, err := io.Copy(sconn1.Conn, sconn2); err != nil {
				log.Println(m.Name, "pipe A error:", err)
			}
		}
		exit <- true
	}()

	go func() {
		log.Println(m.Name, "(Handler1) Iniciando pipe → A", m.Client)
		if m.Client {
			if _, err := io.Copy(sconn2.Conn, sconn1); err != nil {
				log.Println(m.Name, "pipe B error:", err)
			}
		} else {
			if _, err := io.Copy(sconn2, sconn1.Conn); err != nil {
				log.Println(m.Name, "pipe B error:", err)
			}
		}
		exit <- true
	}()

	<-exit // cerramos si cualquiera falla

	log.Println(m.Name, "(Handler1) Cerrando conexiones")
	_ = sconn1.Close()
	_ = sconn2.Close()
	log.Println(m.Name, "(Handler1) Close completo")
}
