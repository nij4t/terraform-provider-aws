package elastictranscoder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccElasticTranscoderPreset_basic(t *testing.T) {
	var preset elastictranscoder.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckElasticTranscoderPresetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(resourceName, &preset),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElasticTranscoderPreset_disappears(t *testing.T) {
	var preset elastictranscoder.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckElasticTranscoderPresetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPresetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(resourceName, &preset),
					testAccCheckElasticTranscoderPresetDisappears(&preset),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Reference: https://github.com/nij4t/terraform-provider-aws/issues/14087
func TestAccElasticTranscoderPreset_AudioCodecOptions_empty(t *testing.T) {
	var preset elastictranscoder.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckElasticTranscoderPresetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPresetAudioCodecOptionsEmptyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(resourceName, &preset),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"audio_codec_options"}, // Due to incorrect schema (should be nested under audio)
			},
		},
	})
}

func TestAccElasticTranscoderPreset_description(t *testing.T) {
	var preset elastictranscoder.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckElasticTranscoderPresetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPresetDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(resourceName, &preset),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Tests all configuration blocks
func TestAccElasticTranscoderPreset_full(t *testing.T) {
	var preset elastictranscoder.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckElasticTranscoderPresetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPresetFull1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(resourceName, &preset),
					resource.TestCheckResourceAttr(resourceName, "audio.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audio_codec_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "thumbnails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "video.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "video_codec_options.%", "5"),
					resource.TestCheckResourceAttr(resourceName, "video_watermarks.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPresetFull2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(resourceName, &preset),
					resource.TestCheckResourceAttr(resourceName, "audio.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "audio_codec_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "thumbnails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "video.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "video_codec_options.%", "5"),
					resource.TestCheckResourceAttr(resourceName, "video_watermarks.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/nij4t/terraform-provider-aws/issues/695
func TestAccElasticTranscoderPreset_Video_frameRate(t *testing.T) {
	var preset elastictranscoder.Preset
	resourceName := "aws_elastictranscoder_preset.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elastictranscoder.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckElasticTranscoderPresetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPresetVideoFrameRateConfig(rName, "29.97"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckElasticTranscoderPresetExists(resourceName, &preset),
					resource.TestCheckResourceAttr(resourceName, "video.0.frame_rate", "29.97"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckElasticTranscoderPresetExists(name string, preset *elastictranscoder.Preset) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderConn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Preset ID is set")
		}

		out, err := conn.ReadPreset(&elastictranscoder.ReadPresetInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*preset = *out.Preset

		return nil
	}
}

func testAccCheckElasticTranscoderPresetDisappears(preset *elastictranscoder.Preset) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderConn
		_, err := conn.DeletePreset(&elastictranscoder.DeletePresetInput{
			Id: preset.Id,
		})

		return err
	}
}

func testAccCheckElasticTranscoderPresetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticTranscoderConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastictranscoder_preset" {
			continue
		}

		out, err := conn.ReadPreset(&elastictranscoder.ReadPresetInput{
			Id: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if out.Preset != nil && *out.Preset.Id == rs.Primary.ID {
				return fmt.Errorf("Elastic Transcoder Preset still exists")
			}
		}

		if !tfawserr.ErrMessageContains(err, elastictranscoder.ErrCodeResourceNotFoundException, "") {
			return fmt.Errorf("unexpected error: %s", err)
		}

	}
	return nil
}

func testAccPresetConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "mp4"
  name      = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 320
    channels           = 2
    codec              = "mp3"
    sample_rate        = 44100
  }
}
`, rName)
}

func testAccPresetAudioCodecOptionsEmptyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "mp4"
  name      = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 320
    channels           = 2
    codec              = "mp3"
    sample_rate        = 44100
  }

  audio_codec_options {}
}
`, rName)
}

func testAccPresetDescriptionConfig(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container   = "mp4"
  description = %[2]q
  name        = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 320
    channels           = 2
    codec              = "mp3"
    sample_rate        = 44100
  }
}
`, rName, description)
}

func testAccPresetFull1Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "mp4"
  name      = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 128
    channels           = 2
    codec              = "AAC"
    sample_rate        = 48000
  }

  audio_codec_options {
    profile = "auto"
  }

  video {
    bit_rate             = "auto"
    codec                = "H.264"
    display_aspect_ratio = "16:9"
    fixed_gop            = "true"
    frame_rate           = "auto"
    keyframes_max_dist   = 90
    max_height           = 1080
    max_width            = 1920
    padding_policy       = "Pad"
    sizing_policy        = "Fit"
  }

  video_codec_options = {
    Profile                  = "main"
    Level                    = "4.1"
    MaxReferenceFrames       = 4
    InterlacedMode           = "Auto"
    ColorSpaceConversionMode = "None"
  }

  thumbnails {
    format         = "jpg"
    interval       = 5
    max_width      = 960
    max_height     = 540
    padding_policy = "Pad"
    sizing_policy  = "Fit"
  }
}
`, rName)
}

func testAccPresetFull2Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "mp4"
  name      = %[1]q

  audio {
    audio_packing_mode = "SingleTrack"
    bit_rate           = 96
    channels           = 2
    codec              = "AAC"
    sample_rate        = 44100
  }

  audio_codec_options {
    profile = "AAC-LC"
  }

  video {
    bit_rate             = "1600"
    codec                = "H.264"
    display_aspect_ratio = "16:9"
    fixed_gop            = "false"
    frame_rate           = "auto"
    max_frame_rate       = "60"
    keyframes_max_dist   = 240
    max_height           = "auto"
    max_width            = "auto"
    padding_policy       = "Pad"
    sizing_policy        = "Fit"
  }

  video_codec_options = {
    Profile                  = "main"
    Level                    = "2.2"
    MaxReferenceFrames       = 3
    InterlacedMode           = "Progressive"
    ColorSpaceConversionMode = "None"
  }

  video_watermarks {
    id                = "Terraform Test"
    max_width         = "20%%"
    max_height        = "20%%"
    sizing_policy     = "ShrinkToFit"
    horizontal_align  = "Right"
    horizontal_offset = "10px"
    vertical_align    = "Bottom"
    vertical_offset   = "10px"
    opacity           = "55.5"
    target            = "Content"
  }

  thumbnails {
    format         = "png"
    interval       = 120
    max_width      = "auto"
    max_height     = "auto"
    padding_policy = "Pad"
    sizing_policy  = "Fit"
  }
}
`, rName)
}

func testAccPresetVideoFrameRateConfig(rName string, frameRate string) string {
	return fmt.Sprintf(`
resource "aws_elastictranscoder_preset" "test" {
  container = "mp4"
  name      = %[1]q

  thumbnails {
    format         = "png"
    interval       = 120
    max_width      = "auto"
    max_height     = "auto"
    padding_policy = "Pad"
    sizing_policy  = "Fit"
  }

  video {
    bit_rate             = "auto"
    codec                = "H.264"
    display_aspect_ratio = "16:9"
    fixed_gop            = "true"
    frame_rate           = %[2]q
    keyframes_max_dist   = 90
    max_height           = 1080
    max_width            = 1920
    padding_policy       = "Pad"
    sizing_policy        = "Fit"
  }

  video_codec_options = {
    Profile                  = "main"
    Level                    = "4.1"
    MaxReferenceFrames       = 4
    InterlacedMode           = "Auto"
    ColorSpaceConversionMode = "None"
  }
}
`, rName, frameRate)
}
