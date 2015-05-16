package main

import (
  "log"
  "io/ioutil"
  "strconv"
  "encoding/xml"
//  "github.com/nfnt/resize"
  "github.com/spf13/cobra"
  "github.com/k0kubun/pp"
)

type Item uint64

type IntegerArray struct {
  Name  string `xml:"name,attr"`
  Items []Item  `xml:"item"`
}

type Resources struct {
  IntegerArrays []IntegerArray `xml:"integer-array"`
}

func (item *Item) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
  var content string
  if err := d.DecodeElement(&content, &start); err != nil {
    return err
  }
  codePoint, err := strconv.ParseUint(content, 0, 64)
  if err != nil {
    return err
  }
  *item = Item(codePoint)
  return nil
}

func main() {
  var Source string
  var SpritegenCmd = &cobra.Command{
    Use: "spritegen",
    Short: "Spritegen takes a set of codepoints and emoji assets and generates sprites for them",
    Run: func(cmd *cobra.Command, args []string) {
      pp.Print(args)
      bytes, err := ioutil.ReadFile("emoji.xml")
      if err != nil {
        log.Fatal(err)
      }
      resources := Resources{}
      err = xml.Unmarshal(bytes, &resources)
      if err != nil {
        log.Fatal(err)
      }
      pp.Print(resources.IntegerArrays)
    },
  }
  SpritegenCmd.Flags().StringVarP(&Source, "input", "i", "emoji.xml", "Source Andriod resource XML file to read from.")
  SpritegenCmd.Execute()
}
