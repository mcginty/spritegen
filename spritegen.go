package main

import (
  "os"
  "log"
  "fmt"
  "io/ioutil"
  "path/filepath"
  "encoding/xml"
//  "github.com/nfnt/resize"
  "github.com/spf13/cobra"
//  "github.com/k0kubun/pp"
)

type CodePoint string

type IntegerArray struct {
  Name       string      `xml:"name,attr"`
  CodePoints []CodePoint `xml:"item"`
}

type Resources struct {
  IntegerArrays []IntegerArray `xml:"integer-array"`
}

func (item *CodePoint) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
  var content string
  if err := d.DecodeElement(&content, &start); err != nil {
    return err
  }
  *item = CodePoint(content[2:])
  return nil
}

func readResources (emojiXml string) (resources *Resources, err error) {
  bytes, err := ioutil.ReadFile(emojiXml)
  if err != nil {
    return nil, err
  }
  resources = &Resources{}
  err = xml.Unmarshal(bytes, resources)
  if err != nil {
    return nil, err
  }
  return
}

func main() {
  var InputXml string
  var EmojiDir string
  var EmojiPrefix string
  var SpritegenCmd = &cobra.Command{
    Use: "spritegen",
    Short: "Spritegen takes a set of codepoints and emoji assets and generates sprites for them",
    Run: func(cmd *cobra.Command, args []string) {
      resources, err := readResources(InputXml)
      if err != nil {
        log.Fatal(err)
      }
      for _, intArray := range resources.IntegerArrays {
        for _, codePoint := range intArray.CodePoints {
          emojipath := filepath.Join(EmojiDir, EmojiPrefix + string(codePoint) + ".png")
          fi, err := os.Stat(emojipath)
          if err != nil {
            log.Fatal(err)
          }
          if fi.Mode().IsRegular() {
            fmt.Printf("%s exists!\n", emojipath)
          } else {
            fmt.Printf("%s DOES NOT EXISTTTT!\n", emojipath)
          }
        }
      }
    },
  }
  SpritegenCmd.Flags().StringVarP(&InputXml, "input", "i", "emoji.xml", "Source Android resource XML file to read from")
  SpritegenCmd.Flags().StringVarP(&EmojiDir, "emoji", "e", "noto/color_emoji/png/128/", "Source emoji folder for lookup")
  SpritegenCmd.Flags().StringVarP(&EmojiPrefix, "emoji-prefix", "p", "emoji_u", "Prefix used by emoji files")
  SpritegenCmd.Execute()
}
