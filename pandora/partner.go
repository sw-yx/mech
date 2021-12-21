package pandora

import (
   "bytes"
   "encoding/hex"
   "encoding/json"
   "github.com/89z/mech"
   "golang.org/x/crypto/blowfish"
   "net/http"
   "strings"
)

const (
   origin = "http://android-tuner.pandora.com"
   partnerPassword = "AC7IBG09A3DTSYM4R41UJWL07VLN8JI7"
   syncTime = 2222222222
)

var (
   LogLevel mech.LogLevel
   key = []byte("6#26FRL$ZWD")
)

func Decrypt(src []byte) ([]byte, error) {
   sLen := len(src)
   if sLen < blowfish.BlockSize {
      return nil, mech.InvalidSlice{blowfish.BlockSize-1, sLen}
   }
   dst := make([]byte, sLen)
   blow, err := blowfish.NewCipher(key)
   if err != nil {
      return nil, err
   }
   for low := 0; low < sLen; low += blowfish.BlockSize {
      blow.Decrypt(dst[low:], src[low:])
   }
   return unpad(dst)
}

func Encrypt(src []byte) ([]byte, error) {
   src = pad(src)
   dst := make([]byte, len(src))
   blow, err := blowfish.NewCipher(key)
   if err != nil {
      return nil, err
   }
   for low := 0; low < len(src); low += blow.BlockSize() {
      blow.Encrypt(dst[low:], src[low:])
   }
   return dst, nil
}

func hexEncode(val interface{}) (string, error) {
   body, err := json.Marshal(val)
   if err != nil {
      return "", err
   }
   buf, err := Encrypt(body)
   if err != nil {
      return "", err
   }
   return hex.EncodeToString(buf), nil
}

func pad(src []byte) []byte {
   sLen := blowfish.BlockSize - len(src) % blowfish.BlockSize
   for high := byte(sLen); sLen >= 1; sLen-- {
      src = append(src, high)
   }
   return src
}

func unpad(src []byte) ([]byte, error) {
   sLen := len(src)
   if sLen == 0 {
      return nil, mech.InvalidSlice{}
   }
   tLen := src[sLen-1]
   high := sLen - int(tLen)
   if high <= -1 {
      return nil, mech.InvalidSlice{high, sLen}
   }
   return src[:high], nil
}

type PartnerLogin struct {
   Result struct {
      PartnerAuthToken string
   }
}

func NewPartnerLogin() (*PartnerLogin, error) {
   body := map[string]string{
      "deviceModel": "android-generic",
      "password": partnerPassword,
      "username": "android",
      "version": "5",
   }
   buf := new(bytes.Buffer)
   err := json.NewEncoder(buf).Encode(body)
   if err != nil {
      return nil, err
   }
   req, err := http.NewRequest("POST", origin + "/services/json/", buf)
   if err != nil {
      return nil, err
   }
   req.URL.RawQuery = "method=auth.partnerLogin"
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   if res.StatusCode != http.StatusOK {
      return nil, mech.Response{res}
   }
   part := new(PartnerLogin)
   if err := json.NewDecoder(res.Body).Decode(part); err != nil {
      return nil, err
   }
   return part, nil
}

func (p PartnerLogin) UserLogin(username, password string) (*UserLogin, error) {
   rUser := userLoginRequest{
      LoginType: "user",
      PartnerAuthToken: p.Result.PartnerAuthToken,
      Password: password,
      SyncTime: syncTime,
      Username: username,
   }
   body, err := hexEncode(rUser)
   if err != nil {
      return nil, err
   }
   req, err := http.NewRequest(
      "POST", origin + "/services/json/", strings.NewReader(body),
   )
   val := make(mech.Values)
   // this can be empty, but must be included:
   val["auth_token"] = ""
   val["method"] = "auth.userLogin"
   val["partner_id"] = "42"
   req.URL.RawQuery = val.Encode()
   LogLevel.Dump(req)
   res, err := new(http.Transport).RoundTrip(req)
   if err != nil {
      return nil, err
   }
   defer res.Body.Close()
   user := new(UserLogin)
   if err := json.NewDecoder(res.Body).Decode(user); err != nil {
      return nil, err
   }
   return user, nil
}

type userLoginRequest struct {
   LoginType string `json:"loginType"`
   PartnerAuthToken string `json:"partnerAuthToken"`
   Password string `json:"password"`
   SyncTime int `json:"syncTime"`
   Username string `json:"username"`
}