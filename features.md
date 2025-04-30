## 🔐 Túnel Seguro Bidireccional (Go)

Este sistema implementa un túnel de red con cifrado extremo a extremo, compresión adaptativa y estructura modular para futuras extensiones.

---

### ✅ Features Implementados

| Feature                            | Estado     | Notas                                                                 |
|-----------------------------------|------------|-----------------------------------------------------------------------|
| 🔐 Cifrado AES-GCM extremo a extremo | ✅ Completo | Todo el tráfico está cifrado punto a punto con framing binario       |
| 🧩 Framing binario + tipos de mensaje | ✅ Completo | `0x02` sin compresión, `0x03` con Zstd, `0x01` para control           |
| 🧬 Compresión condicional (Zstd)     | ✅ Completo | Se aplica solo si la compresión reduce tamaño; configurable por frame|
| 🗝️ Clave compartida desde entorno (`SHARED_KEY`) | ✅ Completo | Validación de longitud, carga segura desde `os.Getenv()`             |
| 🔁 Flujo bidireccional seguro        | ✅ Completo | Manejo simétrico entre cliente y servidor                            |
| 🧪 Trazas básicas y logs informativos | ✅ Parcial  | Logs en consola, con tags por conexión (mejorable con estructurado)  |
| 🔐 Autenticación mutua                | ✅ Completo | Validar identidad deL Cliente (token)       |

---

### 🧩 Features Pendientes

| Prioridad | Feature                              | Estado     | Descripción técnica                                                       |
|-----------|--------------------------------------|------------|---------------------------------------------------------------------------|
| 🟢 Alta   | 🔁 Reconexión automática              | ❌ Pendiente | Detectar cortes y reintentar con backoff exponencial                      |
| 🟢 Alta   | 🫀 Heartbeat / KeepAlive              | ❌ Pendiente | Ping cifrado para detectar caídas silenciosas                             |
| 🟡 Media  | 🔄 Rotación de clave compartida       | ❌ Pendiente | Reacordar `sharedKey` sin interrumpir tráfico                             |
| 🟡 Media  | 🔄 Hot-reload de configuración        | ❌ Pendiente | Cambiar parámetros sin reiniciar                                          |
| 🟡 Media  | 🔐 Cifrado con identidad (opcional)   | ❌ Pendiente | Soporte para TLS/mTLS o Ed25519                                           |
| 🟡 Media  | 🪄 Multiplexing                       | ❌ Pendiente | Varios streams lógicos sobre una única conexión TCP                       |
| 🔵 Baja   | 📈 Métricas Prometheus                | ❌ Pendiente | Exponer tráfico, errores, conexiones                                      |
| 🔵 Baja   | 🪵 Logging estructurado               | ⏳ Parcial  | Niveles, JSON, tags personalizados                                        |
| 🔵 Baja   | 💾 Persistencia del estado            | ❌ Pendiente | Registro de conexiones activas y logs persistentes                        |
| 🔵 Baja   | 🌐 Admin API                          | ❌ Pendiente | HTTP API para monitoreo, reinicio y métricas                              |

---

### 🧪 Enlace de prueba rápida

```bash
SHARED_KEY="thisis32byteslongthisis32byteslo" go run .