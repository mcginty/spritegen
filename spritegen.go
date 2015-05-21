package main

import (
  "os"
  "log"
  "math"
  "io/ioutil"
  "image"
  "image/draw"
  "image/png"
  "path/filepath"
  "encoding/xml"
  "github.com/spf13/cobra"
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

func getSingleSpriteRect(index int, dimens image.Point, pad image.Point, maxWidth int) (rect image.Rectangle) {
  emojiPerLine := maxWidth / (dimens.X + pad.X)
  line := index / emojiPerLine
  row := index % emojiPerLine
  x := row * dimens.X + row * pad.X
  y := line * dimens.Y + line * pad.Y
  rect = image.Rect(x, y, x + dimens.X, y + dimens.Y)
  return rect
}

func getSpriteRect(codePoints []CodePoint, dimens image.Point, pad image.Point, maxWidth int) (rect image.Rectangle) {
  var width        int
  var height       int
  if len(codePoints) * (dimens.X + pad.X) < maxWidth {
    width  = len(codePoints) * dimens.X + (len(codePoints) - 1) * pad.X
    height = dimens.Y
  } else {
    emojiPerLine := maxWidth / (dimens.X + pad.X)
    lines := int(math.Ceil(float64(len(codePoints)) / float64(emojiPerLine)))
    width = emojiPerLine * dimens.X
    height = lines * dimens.Y + pad.Y * (lines - 1)
  }
  rect = image.Rect(0, 0, width, height)
  return
}

func getEmoji(emojipath string) (img image.Image, err error) {
  file, err := os.Open(emojipath)
  if err != nil {
    return nil, err
  }
  defer file.Close()
  img, err = png.Decode(file)
  if err != nil {
    log.Fatal(err)
  }
  return
}

func main() {
  var InputXml string
  var EmojiDir string
  var EmojiPrefix string
  var MaxWidth int
  var EmojiWidth int
  var EmojiHeight int
  var EmojiPadY int
  var EmojiPadX int

  var SpritegenCmd = &cobra.Command{
    Use: "spritegen",
    Short: "Spritegen takes a set of codepoints and emoji assets and generates sprites for them",
    Run: func(cmd *cobra.Command, args []string) {
      resources, err := readResources(InputXml)
      if err != nil {
        log.Fatal(err)
      }
      emojiDimens := image.Pt(EmojiWidth, EmojiHeight)
      padDimens := image.Pt(EmojiPadX, EmojiPadY)
      for _, intArray := range resources.IntegerArrays {
        sprite := image.NewRGBA(getSpriteRect(intArray.CodePoints, emojiDimens, padDimens, MaxWidth))
        for j, codePoint := range intArray.CodePoints {
          emojipath := filepath.Join(EmojiDir, EmojiPrefix + string(codePoint) + ".png")
          m, err := getEmoji(emojipath)
          if err != nil {
            log.Fatal(err)
          }
          draw.Draw(sprite, getSingleSpriteRect(j, emojiDimens, padDimens, MaxWidth), m, image.Pt(0, 0), draw.Over)
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
  SpritegenCmd.Flags().IntVarP(&EmojiWidth, "emoji-width", "x", 136, "Input width of emoji (they are forced square)")
  SpritegenCmd.Flags().IntVarP(&EmojiHeight, "emoji-height", "y", 128, "input height of emoji (they are forced square)")
  SpritegenCmd.Flags().IntVarP(&EmojiPadX, "emoji-pad-x", "q", 0, "Horizontal padding of sprited emoji")
  SpritegenCmd.Flags().IntVarP(&EmojiPadY, "emoji-pad-y", "r", 8, "Vertical padding of sprited emoji")
  SpritegenCmd.Flags().IntVarP(&MaxWidth, "max-width", "m", 2048, "Max width of sprite")
  SpritegenCmd.Execute()
}
