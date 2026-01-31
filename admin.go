package main

// საჭირო პაკეტების შემოტანა
import (
	"bufio" // stdin-იდან და TCP კავშირიდან ტექსტის წასაკითხად
	"fmt"   // ეკრანზე დაბეჭდვისთვის (Print, Println)
	"net"   // TCP კავშირებისთვის (Dial)
	"os"    // ოპერაციულ სისტემასთან სამუშაოდ (stdin)
)

func main() {
	// TCP კავშირის დამყარება SERVER-თან
	// აქ ვუთითებთ server-ის IP-ს და პორტს
	conn, err := net.Dial("tcp", "192.168.56.101:9000") // server IP
	if err != nil {
		// თუ კავშირი ვერ დამყარდა, პროგრამა ავარდება შეცდომით
		panic(err)
	}

	// SERVER-ს ვეუბნებით, რომ ეს კავშირი არის ADMIN
	// server ამით არჩევს ADMIN-ს CLIENT-ისგან
	conn.Write([]byte("ADMIN\n"))

	// ცალკე goroutine, რომელიც მუდმივად კითხულობს პასუხებს SERVER-დან
	// (CLIENT-ის მიერ შესრულებული ბრძანებების შედეგებს)
	go func() {
		// Reader TCP კავშირიდან შემოსული მონაცემებისთვის
		reader := bufio.NewReader(conn)

		for {
			// ვკითხულობთ ერთ ხაზს server-დან
			resp, err := reader.ReadString('\n')
			if err != nil {
				// თუ კავშირი გაწყდა (server/client გაითიშა)
				// goroutine უბრალოდ სრულდება
				return
			}

			// მიღებული პასუხის დაბეჭდვა ტერმინალში
			fmt.Print(resp)
		}
	}()

	// Scanner stdin-დან (კლავიატურიდან) ბრძანებების წასაკითხად
	scanner := bufio.NewScanner(os.Stdin)

	// უსასრულო ციკლი — ADMIN მუდმივად აგზავნის ბრძანებებს
	for {
		// prompt, რომ ვიცოდეთ როდის უნდა ჩავწეროთ ბრძანება
		fmt.Print("admin> ")

		// ვკითხულობთ ერთ ხაზს კლავიატურიდან
		scanner.Scan()

		// scanner.Text() აბრუნებს newline-ის გარეშე ტექსტს,
		// ამიტომ ხელით ვამატებთ "\n"
		cmd := scanner.Text() + "\n"

		// შეყვანილი ბრძანების გაგზავნა SERVER-ზე
		conn.Write([]byte(cmd))
	}
}
