// +build !windows

package v2raypool

import "fmt"

func SetProxy(proxy string) error {
    fmt.Println("------SetProxy--", proxy)
    return nil
}
