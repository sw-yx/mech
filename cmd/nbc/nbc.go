package main

import (
   "errors"
   "fmt"
   "github.com/89z/format"
   "github.com/89z/format/hls"
   "github.com/89z/mech/nbc"
   "io"
   "net/http"
   "os"
   "sort"
)

func newMaster(guid, bandwidth int64, info bool) error {
   access, err := nbc.NewAccessVOD(guid)
   if err != nil {
      return err
   }
   fmt.Println("GET", access.ManifestPath)
   res, err := http.Get(access.ManifestPath)
   if err != nil {
      return err
   }
   defer res.Body.Close()
   master, err := hls.NewScanner(res.Body).Master()
   if err != nil {
      return err
   }
   sort.Slice(master.Streams, func(a, b int) bool {
      return master.Streams[a].Bandwidth < master.Streams[b].Bandwidth
   })
   stream := master.Streams.GetBandwidth(bandwidth)
   if info {
      for _, each := range master.Streams {
         if each.Bandwidth == stream.Bandwidth {
            fmt.Print("!")
         }
         fmt.Println(each)
      }
   } else {
      video, err := nbc.NewVideo(guid)
      if err != nil {
         return err
      }
      return download(stream.RawURI, video.Base())
   }
   return nil
}

func download(addr, base string) error {
   fmt.Println("GET", addr)
   res, err := http.Get(addr)
   if err != nil {
      return err
   }
   defer res.Body.Close()
   seg, err := hls.NewScanner(res.Body).Segment()
   if err != nil {
      return err
   }
   file, err := os.Create(base + hls.TS)
   if err != nil {
      return err
   }
   defer file.Close()
   pro := format.ProgressChunks(file, len(seg.Clear))
   for _, clear := range seg.Clear {
      res, err := http.Get(clear)
      if err != nil {
         return err
      }
      if res.StatusCode != http.StatusOK {
         return errors.New(res.Status)
      }
      pro.AddChunk(res.ContentLength)
      if _, err := io.Copy(pro, res.Body); err != nil {
         return err
      }
      if err := res.Body.Close(); err != nil {
         return err
      }
   }
   return nil
}
