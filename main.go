package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/terraform-exec/tfexec"
)

func main() {
	ctx := context.TODO()
	execPath, err := FindTerraform(ctx)
	if err != nil {
		log.Fatalf("error finding a terraform exectuable: %v", err)
	}

	tf, err := tfexec.NewTerraform(".", execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %v", err)
	}

	r, w := io.Pipe()
	errch := make(chan error)
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("*** %s\n", line)
		}
		if err := scanner.Err(); err != nil {
			errch <- err
		}
		close(errch)
	}()

	if err := tf.ApplyJSON(ctx, w); err != nil {
		log.Fatal(err)
	}
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}
	if err := <-errch; err != nil {
		log.Fatal(err)
	}
}
