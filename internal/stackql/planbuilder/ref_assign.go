package planbuilder

// import (
// 	"fmt"

// 	"github.com/stackql/stackql/internal/stackql/suffix"
// 	"vitess.io/vitess/go/vt/sqlparser"
// )

// func ResolveMethods(tables sqlparser.TableExprs, where *sqlparser.Where) error {
// 	requiredParameters := suffix.NewParameterSuffixMap()
// 	// remainingRequiredParameters := suffix.NewParameterSuffixMap()
// 	optionalParameters := suffix.NewParameterSuffixMap()
// 	for _, tb := range tables {
// 		tbID := tb.GetUniqueId()
// 		method, err := tb.GetMethod()
// 		if err != nil {
// 			return err
// 		}
// 		for k, v := range method.GetRequiredParameters() {
// 			key := fmt.Sprintf("%s.%s", tbID, k)
// 			_, keyExists := requiredParameters.Get(key)
// 			if keyExists {
// 				return fmt.Errorf("key already is required: %s", k)
// 			}
// 			requiredParameters.Put(key, v)
// 		}
// 		for k, vOpt := range method.GetOptionalParameters() {
// 			key := fmt.Sprintf("%s.%s", tbID, k)
// 			_, keyExists := optionalParameters.Get(key)
// 			if keyExists {
// 				return fmt.Errorf("key already is optional: %s", k)
// 			}
// 			optionalParameters.Put(key, vOpt)
// 		}
// 	}
// 	return nil
// }
