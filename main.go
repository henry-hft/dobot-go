package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type response1 struct {
	Success bool `json:"success"`
}
type response2 struct {
	Success  bool       `json:"success"`
	Position [4]float64 `json:"position"`
}

func main() {
	// Verbindung zum Server herstellen
	conn, err := net.Dial("tcp", "192.168.1.6:30004")
	if err != nil {
		fmt.Println("Fehler beim Verbinden (30004):", err)
		return
	}
	defer conn.Close()

	conn2, err := net.Dial("tcp", "192.168.1.6:29999")
	if err != nil {
		fmt.Println("Fehler beim Verbinden (30002):", err)
		return
	}
	defer conn2.Close()

	//commandChannel := make(chan string)

	go func() {
		router := http.NewServeMux()
		router.HandleFunc("/start", func(w http.ResponseWriter, req *http.Request) {
			fmt.Println("EnableRobot()")
			w.Header().Set("Content-Type", "application/json")
			success, _ := sendCommand(conn2, "EnableRobot()")
			response := &response1{
				Success: success,
			}
			jData, _ := json.Marshal(response)
			w.Write(jData)
		})

		router.HandleFunc("/stop", func(w http.ResponseWriter, req *http.Request) {
			fmt.Println("DisableRobot()")
			w.Header().Set("Content-Type", "application/json")
			success, _ := sendCommand(conn2, "DisableRobot()")
			response := &response1{
				Success: success,
			}
			jData, _ := json.Marshal(response)
			w.Write(jData)
		})

		router.HandleFunc("/clear", func(w http.ResponseWriter, req *http.Request) {
			fmt.Println("ClearError()")
			w.Header().Set("Content-Type", "application/json")
			success, _ := sendCommand(conn2, "ClearError()")
			response := &response1{
				Success: success,
			}
			jData, _ := json.Marshal(response)
			w.Write(jData)
		})

		router.HandleFunc("/reset", func(w http.ResponseWriter, req *http.Request) {
			fmt.Println("ResetRobot()")
			w.Header().Set("Content-Type", "application/json")
			success, _ := sendCommand(conn2, "ResetRobot()")
			response := &response1{
				Success: success,
			}
			jData, _ := json.Marshal(response)
			w.Write(jData)
		})

		router.HandleFunc("/open", func(w http.ResponseWriter, req *http.Request) {
			fmt.Println("OpenGripper()")
			w.Header().Set("Content-Type", "application/json")
			success, _ := sendCommand(conn2, "DOExecute(8,1)")
			response := &response1{
				Success: success,
			}
			jData, _ := json.Marshal(response)
			w.Write(jData)
		})

		router.HandleFunc("/close", func(w http.ResponseWriter, req *http.Request) {
			fmt.Println("CloseGripper()")
			w.Header().Set("Content-Type", "application/json")
			success, _ := sendCommand(conn2, "DOExecute(8,0)")
			response := &response1{
				Success: success,
			}
			jData, _ := json.Marshal(response)
			w.Write(jData)
		})

		router.HandleFunc("/position", func(w http.ResponseWriter, req *http.Request) {
			fmt.Println("GetPosition()")
			w.Header().Set("Content-Type", "application/json")
			success, position := sendCommand(conn2, "GetPose()")
			response := &response2{
				Success:  success,
				Position: position,
			}
			jData, _ := json.Marshal(response)
			w.Write(jData)
		})

		router.HandleFunc("/status", func(w http.ResponseWriter, req *http.Request) {
			fmt.Println("RobotMode()")
			success, _ := sendCommand(conn2, "RobotMode()")
			response := &response1{
				Success: success,
			}
			jData, _ := json.Marshal(response)
			w.Write(jData)
		})

		router.HandleFunc("/move/{x}/{y}/{z}/{r}", func(w http.ResponseWriter, req *http.Request) {
			fmt.Println("Move()")

			r, _ := strconv.ParseFloat(req.PathValue("r"), 64)
			w.Header().Set("Content-Type", "application/json")
			response := &response1{}
			if r <= 130 && r >= -210 {
				success, _ := sendCommand(conn2, "MovL("+req.PathValue("x")+","+req.PathValue("y")+","+req.PathValue("z")+","+req.PathValue("r")+",0,0)")
				response = &response1{
					Success: success,
				}
			} else {
				fmt.Println("Verbotener Wert für r")
				response = &response1{
					Success: false,
				}
			}

			jData, _ := json.Marshal(response)
			w.Write(jData)

		})

		server := http.Server{
			Addr:    ":8888",
			Handler: router,
		}

		fmt.Println("Starting server on port :8888")
		server.ListenAndServe()
	}()

	go func() {
		//    lastDataTime := time.Now()
		timeoutDuration := 20 * time.Millisecond

		for {
			buffer := make([]byte, 1024)
			conn.SetReadDeadline(time.Now().Add(timeoutDuration))

			_, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Fehler beim Lesen:", err)
				return
			}
		}
	}()

	for {
		// Hier können weitere Aufgaben im Hauptprozess durchgeführt werden
		// z.B. andere Operationen ausführen oder auf Benutzereingaben warten
		time.Sleep(1 * time.Second)
	}
}

// Funktion, die aufgerufen wird, wenn der Timeout auftritt
func CallMyFunction() {
	fmt.Println("Timeout aufgetreten! Funktion aufgerufen.")
}

func sendCommand(conn net.Conn, message string) (bool, [4]float64) {

	_, write_err := conn.Write([]byte(message))
	if write_err != nil {
		fmt.Println("Fehler: ", write_err)
		return false, [4]float64{}
	}

	reply := make([]byte, 1024)

	_, read_err := conn.Read(reply)
	if read_err != nil {
		fmt.Println("Fehler: ", read_err)
		return false, [4]float64{}
	}
	fmt.Println(string(reply))
	if len(string(reply)) > 0 && string(reply)[0] == '0' {
		fmt.Println("Erfolg")

		regex := regexp.MustCompile(`{([^{}]+)}`)
		matches := regex.FindStringSubmatch(string(reply))

		if len(matches) >= 2 {

			numbers := matches[1]
			numberArray := regexp.MustCompile(`[^,]+`).FindAllString(numbers, -1)

			var floatValues [4]float64
			for i, num := range numberArray[:4] {
				floatVal, err := strconv.ParseFloat(num, 64)
				if err != nil {
					panic(err)
				}
				floatValues[i] = floatVal
			}
			return true, floatValues

		}
		return true, [4]float64{}
	}
	time.Sleep(100 * time.Millisecond)
	return false, [4]float64{}
}
