package admin

import (
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin"
)

type Handler struct {
	c pb.ClientServiceClient
	p pb.ProductServiceClient
	o pb.OrderServiceClient
	r pb.ReportServiceClient
}

func NewHandler(clientService pb.ClientServiceClient, productService pb.ProductServiceClient, orderService pb.OrderServiceClient, reportService pb.ReportServiceClient) *Handler {
	return &Handler{
		c: clientService,
		p: productService,
		o: orderService,
		r: reportService,
	}
}
