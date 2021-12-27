package mech

import (
   "bytes"
   "fmt"
   "io"
   "mime"
   "net/http"
   "net/http/httputil"
   "net/url"
   "os"
   "strconv"
   "strings"
)

func Clean(char rune) rune {
   if strings.ContainsRune(`"*/:<>?\|`, char) {
      return -1
   }
   return char
}

// github.com/golang/go/issues/22318
func ExtensionByType(typ string) (string, error) {
   justType, _, err := mime.ParseMediaType(typ)
   if err != nil {
      return "", err
   }
   switch justType {
   case "audio/mp4":
      return ".m4a", nil
   case "audio/webm":
      return ".weba", nil
   case "video/mp4":
      return ".m4v", nil
   case "video/webm":
      return ".webm", nil
   }
   return "", NotFound{justType}
}

func Truncate(str string, length int) string {
   sLen := len(str)
   if sLen <= 99 {
      return str
   }
   return "..." + str[sLen-99:]
}

type InvalidSlice struct {
   Index, Length int
}

func (i InvalidSlice) Error() string {
   index, length := int64(i.Index), int64(i.Length)
   var buf []byte
   buf = append(buf, "index out of range ["...)
   buf = strconv.AppendInt(buf, index, 10)
   buf = append(buf, "] with length "...)
   buf = strconv.AppendInt(buf, length, 10)
   return string(buf)
}

type LogLevel int

func (l LogLevel) Dump(req *http.Request) error {
   switch l {
   case 0:
      fmt.Println(req.Method, req.URL)
   case 1:
      buf, err := httputil.DumpRequest(req, true)
      if err != nil {
         return err
      }
      os.Stdout.Write(buf)
      if !bytes.HasSuffix(buf, []byte{'\n'}) {
         os.Stdout.WriteString("\n")
      }
   case 2:
      buf, err := httputil.DumpRequestOut(req, true)
      if err != nil {
         return err
      }
      os.Stdout.Write(buf)
   }
   return nil
}

type NotFound struct {
   Input string
}

func (n NotFound) Error() string {
   return strconv.Quote(n.Input) + " not found"
}

type Values map[string]string

func (v Values) Encode() string {
   vals := make(url.Values)
   for key, val := range v {
      vals.Set(key, val)
   }
   return vals.Encode()
}

func (v Values) Header() http.Header {
   vals := make(http.Header)
   for key, val := range v {
      vals.Set(key, val)
   }
   return vals
}

func (v Values) Reader() io.Reader {
   enc := v.Encode()
   return strings.NewReader(enc)
}

func formatFloat(number float64, notation []string) string {
   var exp string
   for _, exp = range notation {
      if number < 1000 {
         break
      }
      number /= 1000
   }
   // no space here, as some number are unitless
   return strconv.FormatFloat(number, 'f', 3, 64) + exp
}

func Format(number float64) string {
   notation := []string{"", " K", " M", " B", " T"}
   return formatFloat(number, notation)
}

func FormatSize(number float64) string {
   notation := []string{" B", " kB", " MB", " GB", " TB"}
   return formatFloat(number, notation)
}

func FormatRate(number float64) string {
   notation := []string{" B/s", " kB/s", " MB/s", " GB/s", " TB/s"}
   return formatFloat(number, notation)
}
