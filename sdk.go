package sdk

//go:generate protoc -I $GOPATH/src/github.com/telecom-tower/towerapi/v1 telecomtower.proto --go_out=plugins=grpc:$GOPATH/src/github.com/telecom-tower/towerapi/v1

import (
	"context"
	"image"
	"image/color"

	"github.com/pkg/errors"
	pb "github.com/telecom-tower/towerapi/v1"
	"google.golang.org/grpc"
)

// PaintMode
var (
	// PaintMode is the default paint mode
	PaintMode = int(pb.PaintMode_PAINT)
	// OverMode paints with alpha channel
	OverMode = int(pb.PaintMode_OVER)
)

var (
	// RollingStop stops autoroll
	RollingStop = int(pb.AutoRoll_STOP)
	// RollingStart sets the layer in autoroll mode
	RollingStart = int(pb.AutoRoll_START)
	// RollingNext schedule a new message
	RollingNext = int(pb.AutoRoll_NEXT)
	// RollingContinue continues the autoroll mode
	RollingContinue = int(pb.AutoRoll_CONTINUE)
)

// Client is the base type for sending drawing commands.
type Client struct {
	client pb.TowerDisplayClient
	stream pb.TowerDisplay_DrawClient
}

// Pixel combines a point and a color.
type Pixel struct {
	Point image.Point
	Color color.Color
}

// NewClient instanciates a new client.
func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		client: pb.NewTowerDisplayClient(conn),
		stream: nil,
	}
}
func colorToPbColor(c color.Color) *pb.Color {
	r, g, b, a := c.RGBA()
	return &pb.Color{
		Red:   r >> 8,
		Green: g >> 8,
		Blue:  b >> 8,
		Alpha: a >> 8,
	}
}

// StartDrawing start a drawing using a context
func (client *Client) StartDrawing(ctx context.Context) error {
	stream, err := client.client.Draw(ctx)
	if err != nil {
		return err
	}
	client.stream = stream
	return nil
}

// Clear clears the given layers.
func (client *Client) Clear(layers ...int) error {
	l := make([]int32, len(layers))
	for i := 0; i < len(layers); i++ {
		l[i] = int32(layers[i])
	}
	return client.stream.Send(&pb.DrawRequest{
		Type: &pb.DrawRequest_Clear{
			Clear: &pb.Clear{
				Layer: l,
			},
		},
	})
}

// SetPixels sets all pixels in a given layer.
func (client *Client) SetPixels(pixels []Pixel, layer int, paintMode int) error {
	px := make([]*pb.Pixel, len(pixels))
	for i, p := range pixels {
		px[i] = &pb.Pixel{
			Point: &pb.Point{
				X: int32(p.Point.X),
				Y: int32(p.Point.Y),
			},
			Color: colorToPbColor(p.Color),
		}
	}
	return client.stream.Send(&pb.DrawRequest{
		Type: &pb.DrawRequest_SetPixels{
			SetPixels: &pb.SetPixels{
				Pixels:    px,
				Layer:     int32(layer),
				PaintMode: pb.PaintMode(paintMode),
			},
		},
	})
}

// DrawRectangle draws a rectangle in a given color on a given layer.
func (client *Client) DrawRectangle(rect image.Rectangle, c color.Color, layer int, paintMode int) error {
	pbc := colorToPbColor(c)
	return client.stream.Send(&pb.DrawRequest{
		Type: &pb.DrawRequest_DrawRectangle{
			DrawRectangle: &pb.DrawRectangle{
				Min: &pb.Point{
					X: int32(rect.Min.X),
					Y: int32(rect.Min.Y),
				},
				Max: &pb.Point{
					X: int32(rect.Max.X),
					Y: int32(rect.Max.Y),
				},
				Color:     pbc,
				Layer:     int32(layer),
				PaintMode: pb.PaintMode(paintMode),
			},
		},
	})
}

// WriteText writes a text.
func (client *Client) WriteText(text string, font string, x int, c color.Color, layer int, paintMode int) error {
	return client.stream.Send(&pb.DrawRequest{
		Type: &pb.DrawRequest_WriteText{
			WriteText: &pb.WriteText{
				Text:      text,
				Font:      font,
				X:         int32(x),
				Color:     colorToPbColor(c),
				Layer:     int32(layer),
				PaintMode: pb.PaintMode(paintMode),
			},
		},
	})
}

// SetLayerOrigin sets the origin of a layer
func (client *Client) SetLayerOrigin(layer int, origin image.Point) error {
	return client.stream.Send(&pb.DrawRequest{
		Type: &pb.DrawRequest_SetLayerOrigin{
			SetLayerOrigin: &pb.SetLayerOrigin{
				Layer: int32(layer),
				Position: &pb.Point{
					X: int32(origin.X),
					Y: int32(origin.Y),
				},
			},
		},
	})
}

// SetLayerAlpha sets the alpha of a layer
func (client *Client) SetLayerAlpha(layer int, alpha int) error {
	return client.stream.Send(&pb.DrawRequest{
		Type: &pb.DrawRequest_SetLayerAlpha{
			SetLayerAlpha: &pb.SetLayerAlpha{
				Layer: int32(layer),
				Alpha: int32(alpha),
			},
		},
	})
}

//AutoRoll sets the autoroll mode of the layer
func (client *Client) AutoRoll(layer int, mode int, entry int, separator int) error {
	return client.stream.Send(&pb.DrawRequest{
		Type: &pb.DrawRequest_AutoRoll{
			AutoRoll: &pb.AutoRoll{
				Layer:     int32(layer),
				Mode:      pb.AutoRoll_Mode(mode),
				Entry:     int32(entry),
				Separator: int32(separator),
			},
		},
	})
}

// Render combines the layers and render the image to the LEDs
func (client *Client) Render() error {
	reply, err := client.stream.CloseAndRecv()
	if err != nil {
		return err
	}
	if reply.Message != "" {
		return errors.New(reply.Message)
	}
	return nil
}
