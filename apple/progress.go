package apple

import (
   "fmt"
   "github.com/89z/mech"
   "net/http"
   "strconv"
)

type Progress struct {
   *http.Response
   callback func(int64, int64)
   x int
   y int64
}

func (p *Progress) Read(buf []byte) (int, error) {
   if p.x == 0 {
      p.callback(p.y, p.ContentLength)
   }
   num, err := p.Body.Read(buf)
   if err != nil {
      return 0, err
   }
   p.y += int64(num)
   p.x += num
   if p.x >= 10_000_000 {
      p.x = 0
   }
   return num, nil
}

// Read method has pointer receiver
func NewProgress(res *http.Response) *Progress {
   var pro Progress
   pro.Response = res
   pro.callback = func(num, den int64) {
      percent := strconv.FormatInt(100*num/den, 10) + "%"
      bytes := mech.FormatSize(float64(num))
      fmt.Println(percent, bytes)
   }
   return &pro
}
