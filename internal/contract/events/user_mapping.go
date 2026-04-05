package events

// Mapping from a producer BC's internal domain events to its Published
// Language contracts (e.g. UserCreatedV1) is performed inside the producing
// BC — see gct/internal/context/iam/user/application/published/. That
// package is allowed to import both its own domain events and this events/
// package. Keeping the mapping there avoids forcing the contracts package
// to know anything about BC internals.
