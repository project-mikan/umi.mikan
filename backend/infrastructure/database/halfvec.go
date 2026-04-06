package database

// Halfvec はpgvectorのhalfvec型をGoで表現するための型
// dbtplが生成するdiaryembedding.dbtpl.goで参照されるため定義する
// 実際のembedding操作はdiary_embeddings.goで[]float32とSQL文字列変換を直接扱う
type Halfvec []float32
