package network

import (
	"encoding/json"
	"log"
)

type ControlMessage struct {
	Type string `json:"type"`
	Port int    `json:"port"`
}

func SendControlMessage(conn *SecureConn, port int) error {

	msg := ControlMessage{
		Type: "control",
		Port: port,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[SendControlMessage] Error al serializar JSON: %v", err)
		return err
	}
	log.Printf("[SendControlMessage] Enviando: %s (%d bytes)", data, len(data))

	_, err = conn.Write(data)
	if err != nil {
		log.Printf("[SendControlMessage] Error al escribir en SecureConn: %v", err)
	}
	return err
}
func ReceiveControlMessage(conn *SecureConn) (ControlMessage, error) {
	var msg ControlMessage
	buf := make([]byte, 1024)

	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("[ReceiveControlMessage] Error al leer: %v", err)
		return msg, err
	}

	log.Printf("[ReceiveControlMessage] Recibido (%d bytes): %x", n, buf[:n])

	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		log.Printf("[ReceiveControlMessage] Error al parsear JSON: %v", err)
	}
	return msg, err
}
