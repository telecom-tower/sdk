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
	// OverMode paints with alpha channel (TODO)
	OverMode = int(pb.PaintMode_OVER)
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

// StartDrawing start a drawing using a context
func (client *Client) StartDrawing(ctx context.Context) error {
	stream, err := client.client.Draw(ctx)
	if err != nil {
		return err
	}
	client.stream = stream
	return nil
}

// Fill fills the layer with the given color.
func (client *Client) Fill(c color.Color, layer int, paintMode int) error {
	r, g, b, a := c.RGBA()
	return client.stream.Send(&pb.DrawRequest{
		Type: &pb.DrawRequest_Fill{
			Fill: &pb.Fill{
				Color: &pb.Color{
					Red:   r >> 8,
					Green: g >> 8,
					Blue:  b >> 8,
					Alpha: a >> 8,
				},
				Layer:     int32(layer),
				PaintMode: pb.PaintMode(paintMode),
			},
		},
	})
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
		r, g, b, a := p.Color.RGBA()
		px[i] = &pb.Pixel{
			Point: &pb.Point{
				X: int32(p.Point.X),
				Y: int32(p.Point.Y),
			},
			Color: &pb.Color{
				Red:   r >> 8,
				Green: g >> 8,
				Blue:  b >> 8,
				Alpha: a >> 8,
			},
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
	r, g, b, a := c.RGBA()
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
				Color: &pb.Color{
					Red:   r >> 8,
					Green: g >> 8,
					Blue:  b >> 8,
					Alpha: a >> 8,
				},
				Layer:     int32(layer),
				PaintMode: pb.PaintMode(paintMode),
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
