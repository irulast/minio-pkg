// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package certs_test

import (
	"context"
	"crypto/tls"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/minio/pkg/v3/certs"
)

func updateCerts(crt, key string) {
	// ignore error handling
	crtSource, _ := os.Open(crt)
	defer crtSource.Close()
	crtDest, _ := os.Create("public.crt")
	defer crtDest.Close()
	io.Copy(crtDest, crtSource)

	keySource, _ := os.Open(key)
	defer keySource.Close()
	keyDest, _ := os.Create("private.key")
	defer keyDest.Close()
	io.Copy(keyDest, keySource)
}

func TestNewManager(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	c, err := certs.NewManager(ctx, "public.crt", "private.key", tls.LoadX509KeyPair)
	if err != nil {
		t.Fatal(err)
	}
	hello := &tls.ClientHelloInfo{}
	gcert, err := c.GetCertificate(hello)
	if err != nil {
		t.Fatal(err)
	}
	expectedCert, err := tls.LoadX509KeyPair("public.crt", "private.key")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gcert.Certificate, expectedCert.Certificate) {
		t.Error("certificate doesn't match expected certificate")
	}
	_, err = certs.NewManager(ctx, "public.crt", "new-private.key", tls.LoadX509KeyPair)
	if err == nil {
		t.Fatal("Expected to fail but got success")
	}

	allCerts := c.GetAllCertificates()
	var found bool
	for _, cert := range allCerts {
		if cert.Issuer.String() != "CN=minio.io,OU=Engineering,O=Minio,L=Redwood City,ST=CA,C=US" {
			t.Error("Unexpected cert issuer found")
		}
		found = true
	}
	if !found {
		t.Error("atleast one public cert is expected")
	}

	var nilMgr *certs.Manager
	if len(nilMgr.GetAllCertificates()) != 0 {
		t.Error("no public cert is expected")
	}
}

func TestValidPairAfterWrite(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	expectedCert, err := tls.LoadX509KeyPair("new-public.crt", "new-private.key")
	if err != nil {
		t.Fatal(err)
	}

	c, err := certs.NewManager(ctx, "public.crt", "private.key", tls.LoadX509KeyPair, func(m *certs.Manager) {
		// Kubernetes uses polling instead of file notifications.
		m.UpdateReloadDuration(25 * time.Millisecond)
	})
	if err != nil {
		t.Fatal(err)
	}

	updateCerts("new-public.crt", "new-private.key")
	defer updateCerts("original-public.crt", "original-private.key")

	deadline := time.Now().Add(5 * time.Second)
	for {
		serverCert, serverErr := c.GetCertificate(&tls.ClientHelloInfo{})
		clientCert, clientErr := c.GetClientCertificate(&tls.CertificateRequestInfo{})
		if serverErr != nil || clientErr != nil {
			t.Fatalf("loading certificates: server=%v client=%v", serverErr, clientErr)
		}
		serverMatches := reflect.DeepEqual(serverCert.Certificate, expectedCert.Certificate)
		clientMatches := reflect.DeepEqual(clientCert.Certificate, expectedCert.Certificate)
		if serverMatches && clientMatches {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("certificate reload did not complete: server=%v client=%v", serverMatches, clientMatches)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func TestNonMatchingCertificate(t *testing.T) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	c, err := certs.NewManager(ctx, "public.crt", "private.key", tls.LoadX509KeyPair)
	if err != nil {
		t.Fatal(err)
	}
	err = c.AddCertificate("server.crt", "server.key")
	if err != nil {
		t.Fatal(err)
	}

	hello := &tls.ClientHelloInfo{ServerName: "non-matching"}
	gcert, err := c.GetCertificate(hello)
	if err != nil {
		t.Fatal(err)
	}
	expectedCert, err := tls.LoadX509KeyPair("public.crt", "private.key")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gcert.Certificate, expectedCert.Certificate) {
		t.Error("certificate doesn't match expected certificate")
	}
	_, err = certs.NewManager(ctx, "public.crt", "new-private.key", tls.LoadX509KeyPair)
	if err == nil {
		t.Fatal("Expected to fail but got success")
	}
}
