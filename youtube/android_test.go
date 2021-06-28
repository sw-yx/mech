package youtube_test

import (
   "github.com/89z/mech/youtube"
   "io"
   "testing"
)

func TestAndroid(t *testing.T) {
   a, err := youtube.NewAndroid("XeojXq6ySs4")
   if err != nil {
      t.Fatal(err)
   }
   if a.Title != "Snowflake" {
      t.Fatalf("%+v\n", a)
   }
   f, err := a.NewFormat(249)
   if err != nil {
      t.Fatal(err)
   }
   if err := f.Write(io.Discard); err != nil {
      t.Fatal(err)
   }
}

func TestFormat(t *testing.T) {
   a, err := youtube.NewAndroid("eAzIAjTBGgU")
   if err != nil {
      t.Fatal(err)
   }
   // this should fail
   f, err := a.NewFormat(302)
   if err == nil {
      t.Fatalf("%+v\n", f)
   }
}
