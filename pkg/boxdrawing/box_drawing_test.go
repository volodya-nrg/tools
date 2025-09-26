package boxdrawing

import (
	"fmt"
	"testing"
)

//nolint:paralleltest,tparallel
func TestBoxDrawing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		title  string
		blocks [][2]string
	}{
		{
			name:  "basic",
			title: "Title",
			blocks: [][2]string{
				{"grpcServer", "http://localhost:8080"},
				{"httpServer", ":808"},
				{"promServer", "127.0.0.1:9090"},
				{"alertManage", "localhost"},
			},
		},
		{
			name:  "only title",
			title: "Title",
		},
		{
			name:  "small 3",
			title: "Title v0.0.0",
			blocks: [][2]string{
				{"x", ""},
				{"y", "1"},
				{"z", "82"},
			},
		},
		{
			name:  "small 2",
			title: "Title v0.0.0",
			blocks: [][2]string{
				{"x", "0"},
				{"y", "1"},
			},
		},
		{
			name:  "small 1",
			title: "Title v0.0.0",
			blocks: [][2]string{
				{"x", "0"},
			},
		},
		{
			name:  "big title",
			title: "Title v0.0.0 lorem ipsum dolor sit amet",
			blocks: [][2]string{
				{"x", "0"},
			},
		},
		{
			name:  "cyrillic",
			title: "Заголовок версии v0.0.",
			blocks: [][2]string{
				{"Прометеус", "http://localhost:9090"},
				{"Джагер", "127.0.0.1:9090"},
				{"Количество ядер", "10"},
				{"Графана", "http://grafana.sait.ru"},
				{"Кол-во горутин", "10"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()

			cfg := NewConfig(2, 1, 3, true).
				WithTitleColor(ColorRed).
				WithOtherColor(ColorPurple)

			bd := NewBoxDrawing(tt.title, cfg)
			for _, block := range tt.blocks {
				bd.AddBlock(block[0], block[1])
			}

			for _, row := range bd.Draw() {
				fmt.Println(row) //nolint:forbidigo
			}
		})
	}
}
