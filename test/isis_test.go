package test

import (
	"isis-total-order-multicast/config"
	"isis-total-order-multicast/server"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func TestThreeNodes(t *testing.T) {
	c1 := config.NodeConfig{
		Id:              "node1",
		Addr:            "127.0.0.1",
		Port:            10001,
		StorageFilename: "node1store.txt",
	}

	s1 := server.NewServer(c1)

	c2 := config.NodeConfig{
		Id:              "node2",
		Addr:            "127.0.0.1",
		Port:            10002,
		StorageFilename: "node2store.txt",
	}

	s2 := server.NewServer(c2)

	c3 := config.NodeConfig{
		Id:              "node3",
		Addr:            "127.0.0.1",
		Port:            10003,
		StorageFilename: "node3store.txt",
	}

	s3 := server.NewServer(c3)

	s1.Start()
	s2.Start()
	s3.Start()

	s1.AddPeer(c1)
	s1.AddPeer(c2)
	s1.AddPeer(c3)

	s2.AddPeer(c1)
	s2.AddPeer(c2)
	s2.AddPeer(c3)

	s3.AddPeer(c1)
	s3.AddPeer(c2)
	s3.AddPeer(c3)

	wg := &sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(3)
		go sendRandomMsg(s1, wg)
		go sendRandomMsg(s2, wg)
		go sendRandomMsg(s3, wg)
		time.Sleep(5 * time.Millisecond)
	}
	wg.Wait()

	outFiles := []string{c1.StorageFilename, c2.StorageFilename, c3.StorageFilename}
	checkOutputFiles(outFiles)
	_ = os.Remove(c1.StorageFilename)
	_ = os.Remove(c2.StorageFilename)
	_ = os.Remove(c3.StorageFilename)
}

func sendRandomMsg(s *server.Server, wg *sync.WaitGroup) {
	s.SendMsg(randStringRunes(8) + "\n")
	wg.Done()
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func checkOutputFiles(outFiles []string) bool {
	if len(outFiles) == 0 {
		return true
	}
	prevDat, err := os.ReadFile(outFiles[0])
	if err != nil {
		panic(err)
	}
	for i := 1; i < len(outFiles); i++ {
		curDat, err := os.ReadFile(outFiles[i])
		if err != nil {
			panic(err)
		}
		if string(curDat) != string(prevDat) {
			return false
		}
		prevDat = curDat
	}
	return true
}
