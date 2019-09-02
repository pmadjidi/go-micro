package main

import (
  "bufio"
  "encoding/base64"
  "encoding/json"
  "flag"
  "fmt"
  "github.com/bitly/go-nsq"
  "log"
  "os"
  "time"
)



type Message struct {
  Inst string `json:"inst"`
  Stamp int64 `json:"stamp"`
  Source string `json:"source"`
  Dest []string `json:"dest"`
  Payload string `json:"payload"`
}





func readImage(imageFile string) string {
  imgfile, err := os.Open(imageFile)
  if err != nil {
    // replace this with real error handling
    panic(err)
  }
  defer imgfile.Close()

  fInfo, _ := imgfile.Stat()
  var size int64 = fInfo.Size()
  buf := make([]byte, size)

  fReader := bufio.NewReader(imgfile)
  fReader.Read(buf)

  imgBase64Str := base64.StdEncoding.EncodeToString(buf)
  return imgBase64Str

}


func main() {
  port := flag.String("port", "4150", "port of the event bus...")
  host := flag.String("host", "localhost", "hostname of the bus...")
  image := flag.String("image", "/Users/payam/Downloads/SampleJPGImage_5mbmb.jpg", "name of image...")
  topic := flag.String("topic", "sts", "topic name for this service...")
  flag.Parse()

  img := readImage(*image)
  config := nsq.NewConfig()
  w, _ := nsq.NewProducer(*host + ":" + *port, config)
  m :=  Message{"Login",time.Now().UnixNano(),"payam",[]string{"Penn", "Teller"},img }
  msg,err := json.Marshal(m)
  if err != nil {
    panic("Error mashaling message.....")
  }
  fmt.Printf("length of message is: %d\n", len(msg))
  err = w.Publish(*topic,msg)
  if err != nil {
      log.Panic("Could not connect")
  }


  w.Stop()
}
