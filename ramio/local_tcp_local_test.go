package ramio

import (
	"testing"
)

func TestLocalToTCPToLocalStream(t *testing.T) {
	return
	// os.MkdirAll("test_data", 0755)
	// inputFile := "test_data/input.txt"
	// outputFile := "test_data/output.txt"
	// defer os.Remove(inputFile)
	// defer os.Remove(outputFile)

	// // Write initial data to input file
	// inputData := []byte("stream integration test")
	// if err := os.WriteFile(inputFile, inputData, 0644); err != nil {
	// 	t.Fatalf("Failed to write input file: %v", err)
	// }

	// // LocalStream for reading
	// localIn := NewLocalStream(ramstream.DRInputStream, inputFile)
	// // LocalStream for writing
	// localOut := NewLocalStream(ramstream.DROutputStream, outputFile)

	// // TCPListener and TCPSender
	// address := "127.0.0.1:9100"
	// tcpStream := NewTCPStream(address)
	// listener := tcpStream
	// sender := tcpStream

	// // Start TCPListener in a goroutine, writing received data to localOut
	// errCh := make(chan error, 1)
	// go func() {
	// 	errCh <- listener.Listen()
	// 	for listener.Len() > 0 {
	// 		buf := make([]byte, listener.Len())
	// 		_, _ = listener.Read(buf)
	// 		_, _ = localOut.Write(buf)
	// 	}
	// }()

	// time.Sleep(100 * time.Millisecond) // Give listener time to start

	// // Read from localIn and send via TCPSender
	// buf := make([]byte, localIn.Len())
	// _, err := localIn.Read(buf)
	// if err != nil {
	// 	t.Fatalf("LocalStream read failed: %v", err)
	// }
	// err = sender.Send(buf)
	// if err != nil {
	// 	t.Fatalf("TCPSender send failed: %v", err)
	// }

	// time.Sleep(200 * time.Millisecond) // Give listener time to process

	// // Verify output file contents
	// outData, err := os.ReadFile(outputFile)
	// if err != nil {
	// 	t.Fatalf("Failed to read output file: %v", err)
	// }
	// if string(outData) != string(inputData) {
	// 	t.Fatalf("Output data mismatch: got %s, want %s", outData, inputData)
	// }
}
