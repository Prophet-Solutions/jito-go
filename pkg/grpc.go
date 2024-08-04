package pkg

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// keepalive client parameters for the gRPC connection
var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send ping every 10 seconds
	Timeout:             5 * time.Second,  // wait 5 seconds for ping ack
	PermitWithoutStream: true,             // send ping even without active streams
}

// CreateGRPCConnection creates and manages a gRPC connection with specified options and error handling.
// It takes a context, an error channel, the gRPC address, and additional gRPC dial options.
// It returns a gRPC client connection or an error if the connection setup fails.
func CreateGRPCConnection(
	ctx context.Context,
	chErr chan error,
	grpcAddr string,
	opts ...grpc.DialOption,
) (*grpc.ClientConn, error) {
	if grpcAddr == "" {
		return nil, errors.New("gRPC address is required")
	}

	// Parse the gRPC address
	u, err := url.Parse(grpcAddr)
	if err != nil {
		return nil, fmt.Errorf("could not parse grpcAddr: %w", err)
	}

	var insecureConnection bool
	if u.Scheme == "http" {
		insecureConnection = true
	}

	// Determine the port based on the scheme
	port := u.Port()
	if port == "" {
		if insecureConnection {
			port = "80"
		} else {
			port = "443"
		}
	}

	hostname := u.Hostname()
	if hostname == "" {
		return nil, errors.New("please provide URL format endpoint e.g. http(s)://<endpoint>:<port>")
	}

	address := hostname + ":" + port

	// Configure transport credentials
	if insecureConnection {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		pool, _ := x509.SystemCertPool()
		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	opts = append(opts, grpc.WithKeepaliveParams(kacp))

	// Create the gRPC client connection
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not create gRPC connection: %w", err)
	}

	// Monitor the connection state in a goroutine
	go func() {
		var retries int
		for {
			select {
			case <-ctx.Done():
				if err = conn.Close(); err != nil {
					chErr <- err
				}
				return
			default:
				state := conn.GetState()
				if state == connectivity.Ready {
					retries = 0
					time.Sleep(1 * time.Second)
					continue
				}

				// Handle different connection states and retries
				if state == connectivity.TransientFailure ||
					state == connectivity.Connecting ||
					state == connectivity.Idle {
					if retries < 5 {
						time.Sleep(time.Duration(retries) * time.Second)
						conn.ResetConnectBackoff()
						retries++
					} else {
						conn.Close()
						conn, err = grpc.DialContext(ctx, address, opts...)
						if err != nil {
							chErr <- err
						}
						retries = 0
					}
				} else if state == connectivity.Shutdown {
					conn, err = grpc.DialContext(ctx, address, opts...)
					if err != nil {
						chErr <- err
					}
					retries = 0
				}

				// Wait for a state change in the connection
				if !conn.WaitForStateChange(ctx, state) {
					continue
				}
			}
		}
	}()

	return conn, nil
}
