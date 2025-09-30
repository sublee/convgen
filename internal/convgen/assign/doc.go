package assign

// Error-Returning Flags
// =====================
//
// - conv.Func.HasErr:
//   Determined by the injector signature defined by the user. If true, the
//   final generated code will return an error regardless of whether the
//   underlying assigner itself may return one.
//
// - factory.allowsErr:
//   Indicates, during assigner building, whether the assigner is allowed to
//   return an error. If an assigner should return an error but this flag is
//   false, the build will fail.
//
// - assigner.requiresErr:
//   After an assigner is built, specifies whether it actually needs to return
//   an error due to underlying operations. It can be false even if the
//   outermost converter returns an error. However, it cannot be true if the
//   outermost converter does not return an error.
//
// - varErr:
//   In writeAssignCode, the variable name used to hold an error, if any. Serves
//   as the propagation mechanism allowing underlying assigners to determine
//   whether they should set the error variable.
