// Package ports hosts consumer-defined interfaces (Anti-Corruption Layer).
// A consumer BC declares the shape of data/operations it needs from a
// supplier BC; the supplier writes an adapter and the composition root
// wires it in. Neither BC imports the other.
package ports
