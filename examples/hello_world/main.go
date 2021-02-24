package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gu827356/pb-gin/examples/hello_world/temp"
)

func main() {
	engine := gin.Default()
	temp.RegisterHelloWorldServiceServer(engine, &Server{})

	err := engine.Run("localhost:8888")
	if err != nil {
		panic(err)
	}
}

type Server struct {
}

func (_ *Server) Hi(_ *gin.Context, in *temp.HiReq) (*temp.HiResp, error) {
	return &temp.HiResp{
		Msg: fmt.Sprintf("name=%s, id=%d, age=%d", in.Name, in.Id, in.Age),
	}, nil
}
