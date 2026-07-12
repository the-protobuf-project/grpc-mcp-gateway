package runtime

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
)

func TestInvokeUnary_NilInterceptorCallsHandler(t *testing.T) {
	called := false
	out, err := InvokeUnary(context.Background(), nil, "/pkg.Svc/M", "srv", "req",
		func(ctx context.Context, req any) (any, error) {
			called = true
			if req != "req" {
				t.Errorf("req = %v, want %q", req, "req")
			}
			return "resp", nil
		})
	if err != nil || out != "resp" || !called {
		t.Fatalf("out=%v err=%v called=%t", out, err, called)
	}
}

func TestInvokeUnary_InterceptorWrapsWithInfo(t *testing.T) {
	interceptor := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if info.FullMethod != "/pkg.Svc/M" {
			t.Errorf("FullMethod = %q", info.FullMethod)
		}
		if info.Server != "srv" {
			t.Errorf("Server = %v", info.Server)
		}
		if req == "reject" {
			return nil, errors.New("rejected")
		}
		return handler(ctx, req)
	}

	if _, err := InvokeUnary(context.Background(), interceptor, "/pkg.Svc/M", "srv", "reject", nil); err == nil {
		t.Fatal("want interceptor rejection, got nil error")
	}
	out, err := InvokeUnary(context.Background(), interceptor, "/pkg.Svc/M", "srv", "ok",
		func(ctx context.Context, req any) (any, error) { return "resp", nil })
	if err != nil || out != "resp" {
		t.Fatalf("out=%v err=%v", out, err)
	}
}

// The chain must run in registration order (first outermost), like
// grpc.ChainUnaryInterceptor.
func TestChainUnaryInterceptors_OrderAndNilHandling(t *testing.T) {
	if ChainUnaryInterceptors() != nil {
		t.Fatal("empty chain should be nil")
	}
	if ChainUnaryInterceptors(nil, nil) != nil {
		t.Fatal("all-nil chain should be nil")
	}

	var order []string
	mk := func(name string) grpc.UnaryServerInterceptor {
		return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			order = append(order, name+":in")
			out, err := handler(ctx, req)
			order = append(order, name+":out")
			return out, err
		}
	}
	chain := ChainUnaryInterceptors(mk("a"), nil, mk("b"))
	out, err := InvokeUnary(context.Background(), chain, "/pkg.Svc/M", nil, "req",
		func(ctx context.Context, req any) (any, error) {
			order = append(order, "handler")
			return "resp", nil
		})
	if err != nil || out != "resp" {
		t.Fatalf("out=%v err=%v", out, err)
	}
	want := []string{"a:in", "b:in", "handler", "b:out", "a:out"}
	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}
