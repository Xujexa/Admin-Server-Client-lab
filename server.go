package main

// საჭირო პაკეტების შემოტანა
import (
	"bufio" // TCP კავშირიდან ტექსტის ხაზ-ხაზად წასაკითხად
	"fmt"   // კონსოლში ინფორმაციის გამოსატანად
	"net"   // TCP სერვერისა და კავშირებისთვის
)

// გლობალური ცვლადები, სადაც ვინახავთ კავშირებს
// adminConn — ADMIN პროგრამის TCP კავშირი
// clientConn — CLIENT პროგრამის TCP კავშირი
var adminConn net.Conn
var clientConn net.Conn

func main() {
	// TCP listener-ის შექმნა პორტზე 9000
	// ეს არის სერვერის „მსმენელი“
	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		// თუ პორტის გახსნა ვერ მოხერხდა — პროგრამა შეწყდება
		panic(err)
	}

	// ინფორმაციის დაბეჭდვა, რომ ვიცოდეთ სერვერი მუშაობს
	fmt.Println("[+] Server listening on :9000")

	// უსასრულო ციკლი — სერვერი მუდმივად ელოდება ახალ კავშირებს
	for {
		// Accept() ელოდება შემოსულ TCP კავშირს
		conn, _ := ln.Accept()

		// თითოეული ახალი კავშირი ცალკე goroutine-ში მუშავდება
		// რომ სერვერი ერთ კავშირზე არ გაიჭედოს
		go identify(conn)
	}
}

func identify(conn net.Conn) {
	// Reader TCP კავშირიდან პირველი ხაზის წასაკითხად
	reader := bufio.NewReader(conn)

	// პირველი ხაზით ვიგებთ ვინ გვიერთდება: ADMIN თუ CLIENT
	role, err := reader.ReadString('\n')
	if err != nil {
		// თუ კავშირი მაშინვე გაწყდა — ფუნქცია სრულდება
		return
	}

	// თუ კლიენტმა გამოაგზავნა "ADMIN\n"
	if role == "ADMIN\n" {
		// ვინახავთ adminConn-ში
		adminConn = conn
		fmt.Println("[+] Admin connected")

		// თუ კლიენტმა გამოაგზავნა "CLIENT\n"
	} else if role == "CLIENT\n" {
		// ვინახავთ clientConn-ში
		clientConn = conn
		fmt.Println("[+] Client connected")
	}

	// თუ ორივე მხარე უკვე დაკავშირებულია,
	// ვიწყებთ მონაცემების გადაგზავნას მათ შორის
	if adminConn != nil && clientConn != nil {
		// ADMIN → CLIENT (ბრძანებების გაგზავნა)
		go pipe(adminConn, clientConn)

		// CLIENT → ADMIN (ბრძანებების შედეგების დაბრუნება)
		go pipe(clientConn, adminConn)
	}
}

func pipe(src net.Conn, dst net.Conn) {
	// Reader წყარო კავშირიდან (src)
	reader := bufio.NewReader(src)

	// უსასრულო ციკლი — სანამ კავშირი არსებობს
	for {
		// ვკითხულობთ ერთ ხაზს წყარო კავშირიდან
		data, err := reader.ReadString('\n')
		if err != nil {
			// თუ კავშირი გაწყდა (ADMIN ან CLIENT გაითიშა)
			// pipe ფუნქცია სრულდება
			return
		}

		// მიღებული მონაცემის გადაგზავნა დანიშნულ კავშირზე
		dst.Write([]byte(data))
	}
}
