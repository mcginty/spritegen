package main

import (
  "os"
  "log"
  "errors"
  "math"
  "io/ioutil"
  "image"
  "image/draw"
  "image/png"
  "path/filepath"
  "encoding/xml"
  "github.com/nfnt/resize"
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

func getSingleSpriteRect(index int, emojiWidth int, maxWidth int) (image.Rectangle) {
  emojiPerLine := maxWidth / emojiWidth
  x := (index % emojiPerLine) * emojiWidth
  y := index / emojiPerLine * emojiWidth
  return image.Rect(x, y, x + emojiWidth, y + emojiWidth)
}

func getSpriteRect(codePoints []CodePoint, emojiWidth int, maxWidth int) (rect image.Rectangle) {
  var width        int
  var height       int
  if len(codePoints) * emojiWidth < maxWidth {
    width  = len(codePoints) * emojiWidth
    height = emojiWidth
  } else {
    emojiPerLine := maxWidth / emojiWidth
    width = emojiPerLine * emojiWidth
    height = int(math.Ceil(float64(len(codePoints)) / float64(emojiPerLine))) * emojiWidth
  }
  rect = image.Rect(0, 0, width, height)
  return
}

func getResizedEmoji(emojipath string, emojiWidth int) (img image.Image, err error) {
  fi, err := os.Stat(emojipath)
  if err != nil {
    return nil, err
  }
  if !fi.Mode().IsRegular() {
    return nil, errors.New("file doesn't exist")
  }
  file, err := os.Open(emojipath)
  if err != nil {
    return nil, err
  }
  defer file.Close()
  rawImage, err := png.Decode(file)
  if err != nil {
    log.Fatal(err)
  }
  img = resize.Resize(uint(emojiWidth), uint(emojiWidth), rawImage, resize.Lanczos3)
  return
}

func main() {
  var InputXml string
  var EmojiDir string
  var EmojiPrefix string
  var MaxWidth int
  var EmojiDimen int
  var SpritegenCmd = &cobra.Command{
    Use: "spritegen",
    Short: "Spritegen takes a set of codepoints and emoji assets and generates sprites for them",
    Run: func(cmd *cobra.Command, args []string) {
      resources, err := readResources(InputXml)
      if err != nil {
        log.Fatal(err)
      }
      for _, intArray := range resources.IntegerArrays {
        sprite := image.NewRGBA(getSpriteRect(intArray.CodePoints, EmojiDimen, MaxWidth))
        for j, codePoint := range intArray.CodePoints {
          emojipath := filepath.Join(EmojiDir, EmojiPrefix + string(codePoint) + ".png")
          m, err := getResizedEmoji(emojipath, EmojiDimen)
          if err != nil {
            log.Fatal(err)
          }
          draw.Draw(sprite, getSingleSpriteRect(j, EmojiDimen, MaxWidth), m, image.Pt(0, 0), draw.Over)
        }
        out, err := os.Create(intArray.Name + ".png")
        if err != nil {
          log.Fatal(err)
        }
        png.Encode(out, sprite)
        out.Close()
      }
    },
  }
  SpritegenCmd.Flags().StringVarP(&InputXml, "input", "i", "emoji.xml", "Source Android resource XML file to read from")
  SpritegenCmd.Flags().StringVarP(&EmojiDir, "emoji", "e", "noto/color_emoji/png/128/", "Source emoji folder for lookup")
  SpritegenCmd.Flags().StringVarP(&EmojiPrefix, "emoji-prefix", "p", "emoji_u", "Prefix used by emoji files")
  SpritegenCmd.Flags().IntVarP(&EmojiDimen, "emoji-dimen", "d", 128, "Target width/height of emoji (they are forced square)")
  SpritegenCmd.Flags().IntVarP(&MaxWidth, "max-width", "w", 2048, "Max width of sprite")
  SpritegenCmd.Execute()
}
