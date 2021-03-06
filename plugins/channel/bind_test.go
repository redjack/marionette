package channel_test

import (
	"context"
	"errors"
	"testing"

	"github.com/redjack/marionette"
	"github.com/redjack/marionette/mock"
	"github.com/redjack/marionette/plugins/channel"
	"go.uber.org/zap"
)

func init() {
	if !testing.Verbose() {
		marionette.Logger = zap.NewNop()
	}
}

func TestBind(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.VarFn = func(name string) interface{} { return nil }
		fsm.ListenFn = func() (int, error) { return 54321, nil }

		var setVarInvoked bool
		fsm.SetVarFn = func(name string, value interface{}) {
			setVarInvoked = true
			if name != "ftp_pasv_port" {
				t.Fatalf("unexpected name: %s", name)
			} else if value != 54321 {
				t.Fatalf("unexpected value: %d", value)
			}
		}

		if err := channel.Bind(context.Background(), &fsm, "ftp_pasv_port"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrNotEnoughArguments", func(t *testing.T) {
		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := channel.Bind(context.Background(), &fsm); err == nil || err.Error() != `not enough arguments` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	t.Run("ErrInvalidArgument", func(t *testing.T) {
		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		if err := channel.Bind(context.Background(), &fsm, 123); err == nil || err.Error() != `invalid argument type` {
			t.Fatalf("unexpected error: %q", err)
		}
	})

	// Ensure plugin is no-op if port already set.
	t.Run("AlreadyBound", func(t *testing.T) {
		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.VarFn = func(name string) interface{} { return 54321 }

		if err := channel.Bind(context.Background(), &fsm, "ftp_pasv_port"); err != nil {
			t.Fatal(err)
		}
	})

	// Ensure error is passed through if listen fails.
	t.Run("ErrListen", func(t *testing.T) {
		errMarker := errors.New("marker")
		conn := mock.DefaultConn()
		fsm := mock.NewFSM(&conn, marionette.NewStreamSet())
		fsm.PartyFn = func() string { return marionette.PartyClient }
		fsm.VarFn = func(name string) interface{} { return nil }
		fsm.ListenFn = func() (int, error) { return 0, errMarker }

		if err := channel.Bind(context.Background(), &fsm, "ftp_pasv_port"); err != errMarker {
			t.Fatal(err)
		}
	})
}
