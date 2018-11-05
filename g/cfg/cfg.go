// cfg.go: configuration handling for cncenter programs

// It expects configuration in 'config' directory near executable
// there must be config/config.default, alternativelly config/config.${APP_ENV}
// code replaces $ENVIRONMENT ${VARIABLES} by the content of environment
// Example:
// # config/config.default
//	db_url: postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@192.168.1.1:5432
//  verbose: false
//  jobs: 14
package cfg

import (
	"fmt"
	"os"
	"reflect"

	yaml "gopkg.in/yaml.v2"
)

type ErrCfg string

func (e ErrCfg) Error() string {
	return fmt.Sprintf("%s", string(e))
}

func file_exists(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}

	return fi.Mode().IsRegular()
}

// get location of config files or error
func Paths() ([]string, error) {

	ret := make([]string, 0, 2)

	if !file_exists("config/config.default") {
		return nil, ErrCfg("'config/config.default' does not exists")
	}
	ret = append(ret, "config/config.default")

	val, ok := os.LookupEnv("APP_ENV")
	if !ok {
		return ret, nil
	}
	env_file := fmt.Sprintf("config/config.%s", val)

	if !file_exists(env_file) {
		return nil, ErrCfg(fmt.Sprintf("'%s' does not exists, check your APP_ENV value", env_file))
	}
	ret = append(ret, env_file)

	return ret, nil
}

//	expand environment variables of all string fields
//	panic if wrong data are supplied in (nil, non struct, ...)
func expand_var(val reflect.Value) {
	for i := 0; i < val.NumField(); i++ {
		f := val.Type().Field(i)
		if f.Type.Kind() == reflect.String {
			v := val.FieldByIndex(f.Index)
			v.SetString(os.ExpandEnv(v.String()))
		} else if (f.Type.Kind() == reflect.Struct) {
			expand_var(val.FieldByIndex(f.Index))
		}
	}
}

// Load conf to given struct
// panic if wrong data are supplied in (nil, non struct, ...)
func Load(out interface{}) error {

	paths, err := Paths()
	if err != nil {
		return err
	}

	for _, p := range paths {
		cf, err := os.Open(p)
		if err != nil {
			return err
		}
		defer cf.Close()
		err = yaml.NewDecoder(cf).Decode(out)
		if err != nil {
			return err
		}
	}
	val := reflect.ValueOf(out).Elem()
	expand_var(val)

	return nil
}
