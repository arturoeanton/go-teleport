package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

type Mirror struct {
	Name              string        `json:"name"`
	Protocol          string        `json:"protocol"`
	Addr1             string        `json:"addr1"`
	Addr2             string        `json:"addr2"`
	InOut1            string        `json:"in_out1"`
	InOut2            string        `json:"in_out2"`
	Conn1             *net.Conn     `json:"-"`
	Conn2             *net.Conn     `json:"-"`
	ChannelNewConn1   chan net.Conn `json:"-"`
	ChannelNewConn2   chan net.Conn `json:"-"`
	channelNewConnCmd chan net.Conn `json:"-"`
	ChannelEventExit  chan bool     `json:"-"`
	channelEventExit1 chan bool     `json:"-"`
	channelEventExit2 chan bool     `json:"-"`
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

func ConnectTo(protocol, addr string) (*net.Conn, error) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		log.Println("(ConnectTo)", err.Error())
		return nil, err
	}
	return &conn, nil
}

func AcceptConn(network, addr string, channelNewConn chan net.Conn, channelEventExit chan bool) {
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
			channelNewConn <- conn
		}
	}()
	<-channelEventExit
	fmt.Println("(AcceptConn) Exit")
}

func (m *Mirror) Start() {
	go m.Handler()

	m.ChannelNewConn1 = make(chan net.Conn)
	m.ChannelNewConn2 = make(chan net.Conn)
	m.channelNewConnCmd = make(chan net.Conn)
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
		fmt.Println(m.Name, "(Start) CMD ConnectTo", m.Addr1)
		m.Conn1, err = ConnectTo(m.Protocol, m.Addr1)
		if err != nil {
			log.Println(m.Name, "(Handler Conn1)", err.Error())
			return
		}
		m.channelNewConnCmd <- *m.Conn1
	}

}

func (m *Mirror) Handler() {
	fmt.Println("(Handler) Start")
	for {
		select {
		case connCmd := <-m.channelNewConnCmd:
			fmt.Println(m.Name, "(Handler) New ConnCMD")
			go m.HandlerCmd(connCmd)
		case conn1 := <-m.ChannelNewConn1:
			fmt.Println(m.Name, "(Handler) New Conn1")
			go m.Handler1(conn1)

		case conn2 := <-m.ChannelNewConn2:
			m.Conn2 = &conn2
			fmt.Println(m.Name, "(Handler) New Conn2")

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

func (m *Mirror) HandlerCmd(conn1 net.Conn) {
	for {
		buf := make([]byte, 1024)
		nbytes, err := conn1.Read(buf)
		if err != nil {
			fmt.Println(m.Name, "(HandlerCmd) Read", err.Error())
			return
		}
		addr := conn1.RemoteAddr().String()
		ip := strings.Split(addr, ":")[0]
		fmt.Println(m.Name, "(HandlerCmd) Read", ip+string(buf[:nbytes]))
		conn, err := net.Dial(m.Protocol, ip+string(buf[:nbytes]))
		if err != nil {
			fmt.Println(m.Name, "(HandlerCmd) Dial", err.Error())
			return
		}
		m.ChannelNewConn1 <- conn
	}
}

func (m *Mirror) Handler1(conn1 net.Conn) {
	fmt.Println(m.Name, "(Handler) Start - ", m.Name)
	var err error
	var conn2 net.Conn
	//conn1 := *m.Conn1

	if m.Conn2 != nil {
		conn2 = *m.Conn2
	}

	if m.InOut2 == "out" {
		conn2, err = net.Dial(m.Protocol, m.Addr2)
		if err != nil {
			log.Println(m.Name, "(Handler1)", err.Error())
			return
		}
	} else {
		port, _ := GetFreePort()
		fmt.Println(m.Name, "(Handler1) port:", port)

		(*m.Conn2).Write([]byte(":" + strconv.Itoa(port))) // deberia ser json y structurarlo

		listener, _ := net.Listen(m.Protocol, ":"+strconv.Itoa(port))
		fmt.Println(m.Name, "(Handler1) listener ... :"+strconv.Itoa(port))
		conn2, err = listener.Accept()
		if err != nil {
			log.Println(m.Name, "(Handler1)(Pasive)", err.Error())
			return
		}
		defer conn2.Close()

	}

	if conn2 == nil {
		log.Println(m.Name, "(Handler1) conn2 is nil")
		return
	}

	go func() {
		if _, err := io.Copy(conn1, conn2); err != nil {
			log.Println(m.Name, "01", err.Error())
			conn1.Close()
			conn2.Close()
			return
		}
	}()
	if _, err := io.Copy(conn2, conn1); err != nil {
		log.Println(m.Name, "02", err.Error())
	}
	conn1.Close()
	conn2.Close()
	fmt.Println(m.Name, "(Handler1) Exit")
}
