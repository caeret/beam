/**
 * @Author: Hao
 * @Date: 17/8/9 14:47
 */

package beam

import (
	"strings"
)

var (
	crlf = []byte{'\r', '\n'}
)

func escapeCrlf(data string) string {
	return strings.NewReplacer("\r", "\\r", "\n", "\\n").Replace(data)
}
