//go:build !solution

package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

func MakeHandler(obj interface{}) http.Handler {
	mux := http.NewServeMux()
	objT := reflect.TypeOf(obj)
	for i := 0; i < objT.NumMethod(); i++ {
		m := objT.Method(i)
		err := registerHandlerFunc(obj, m, mux)
		if err != nil {
			continue
		}
	}
	return mux
}

func registerHandlerFunc(obj interface{}, m reflect.Method, mux *http.ServeMux) error {
	mn := m.Name
	f := m.Func
	mt := m.Type
	inputType, err := methodInputType(mt)
	if err != nil {
		return fmt.Errorf("while register %v method: %w", mn, err)
	}

	mux.Handle("/"+mn, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		inValue := reflect.New(inputType)
		in := inValue.Interface()
		err = json.Unmarshal(bytes, in)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		ctx := r.Context()
		inValues := []reflect.Value{reflect.ValueOf(obj), reflect.ValueOf(ctx), inValue}
		outValues := f.Call(inValues)
		out := outValues[0]
		errV := outValues[1]
		if !errV.IsNil() {
			err = errV.Interface().(error)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		outBytes, _ := json.Marshal(out.Interface())
		w.WriteHeader(http.StatusOK)
		w.Write(outBytes)
	}))
	return nil
}

func methodInputType(mt reflect.Type) (inType reflect.Type, err error) {
	numIns := mt.NumIn()
	if numIns < 3 {
		return (reflect.Type)(nil), fmt.Errorf("not enough input arguments")
	}
	paramType := mt.In(2)
	if paramType.Kind() != reflect.Ptr {
		return (reflect.Type)(nil), fmt.Errorf("third parameter must be a pointer")
	}
	return paramType.Elem(), nil
}

func Call(ctx context.Context, endpoint string, method string, req, rsp interface{}) error {
	client := &http.Client{}
	requestURL := fmt.Sprintf("%s/%s", endpoint, method)
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(req)
	if err != nil {
		return fmt.Errorf("error while encoding input struct: %w", err)
	}

	post, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, &buf)
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	response, err := client.Do(post)
	if err != nil {
		return fmt.Errorf("error while making request from client: %w", err)
	}

	defer response.Body.Close()
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error while reading response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("server error: %s", string(responseBytes))
	}

	return json.Unmarshal(responseBytes, &rsp)
}
