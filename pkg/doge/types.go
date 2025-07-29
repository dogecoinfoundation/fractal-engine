package doge

import "github.com/shopspring/decimal"

type BlockchainInfo struct {
	Chain                string  `json:"chain"`                // (string) current network name (main, test, regtest)
	Blocks               int64   `json:"blocks"`               // (numeric) the height of the most-work fully-validated chain. The genesis block has height 0
	Headers              int64   `json:"headers"`              // (numeric) the current number of headers we have validated
	BestBlockHash        string  `json:"bestblockhash"`        // (string) the hash of the currently best block
	Difficulty           float64 `json:"difficulty"`           // (numeric) the current difficulty
	MedianTime           int64   `json:"mediantime"`           // (numeric) median time for the current best block
	VerificationProgress float64 `json:"verificationprogress"` // (numeric) estimate of verification progress [0..1]
	InitialBlockDownload bool    `json:"initialblockdownload"` // (boolean) (debug information) estimate of whether this node is in Initial Block Download mode
	ChainWord            string  `json:"chainwork"`            // (string) total amount of work in active chain, in hexadecimal
	SizeOnDisk           int64   `json:"size_on_disk"`         // (numeric) the estimated size of the block and undo files on disk
	Pruned               bool    `json:"pruned"`               // (boolean) if the blocks are subject to pruning
	PruneHeight          int64   `json:"pruneheight"`          // (numeric) lowest-height complete block stored (only present if pruning is enabled)
	AutomaticPruning     bool    `json:"automatic_pruning"`    // (boolean) whether automatic pruning is enabled (only present if pruning is enabled)
	PruneTargetSize      int64   `json:"prune_target_size"`    // (numeric) the target size used by pruning (only present if automatic pruning is enabled)
}

type Block struct {
	Hash              string          `json:"hash"`              // (string) the block hash (same as provided) (hex)
	Confirmations     int64           `json:"confirmations"`     // (numeric) The number of confirmations, or -1 if the block is not on the main chain
	Size              int             `json:"size"`              // (numeric) The block size
	StrippedSize      int             `json:"strippedsize"`      // (numeric) The block size excluding witness data
	Weight            int             `json:"weight"`            // (numeric) The block weight as defined in BIP 141
	Height            int64           `json:"height"`            // (numeric) The block height or index
	Version           int             `json:"version"`           // (numeric) The block version
	VersionHex        string          `json:"versionHex"`        // (string) The block version formatted in hexadecimal
	MerkleRoot        string          `json:"merkleroot"`        // (string) The merkle root (hex)
	Time              int             `json:"time"`              // (numeric) The block time in seconds since UNIX epoch (Jan 1 1970 GMT)
	MedianTime        int             `json:"mediantime"`        // (numeric) The median block time in seconds since UNIX epoch (Jan 1 1970 GMT)
	Nonce             int             `json:"nonce"`             // (numeric) The nonce
	Bits              string          `json:"bits"`              // (string) The bits
	Difficulty        decimal.Decimal `json:"difficulty"`        // (numeric) The difficulty
	ChainWork         string          `json:"chainwork"`         // (string) Expected number of hashes required to produce the chain up to this block (hex)
	PreviousBlockHash string          `json:"previousblockhash"` // (string) The hash of the previous block (hex)
	NextBlockHash     string          `json:"nextblockhash"`     // (string) The hash of the next block (hex)
	Tx                []string        `json:"tx"`                // (json array) The transaction ids
}

type BlockWithTransactions struct {
	Hash              string          `json:"hash"`              // (string) the block hash (same as provided) (hex)
	Confirmations     int64           `json:"confirmations"`     // (numeric) The number of confirmations, or -1 if the block is not on the main chain
	Size              int             `json:"size"`              // (numeric) The block size
	StrippedSize      int             `json:"strippedsize"`      // (numeric) The block size excluding witness data
	Weight            int             `json:"weight"`            // (numeric) The block weight as defined in BIP 141
	Height            int64           `json:"height"`            // (numeric) The block height or index
	Version           int             `json:"version"`           // (numeric) The block version
	VersionHex        string          `json:"versionHex"`        // (string) The block version formatted in hexadecimal
	MerkleRoot        string          `json:"merkleroot"`        // (string) The merkle root (hex)
	Time              int             `json:"time"`              // (numeric) The block time in seconds since UNIX epoch (Jan 1 1970 GMT)
	MedianTime        int             `json:"mediantime"`        // (numeric) The median block time in seconds since UNIX epoch (Jan 1 1970 GMT)
	Nonce             int             `json:"nonce"`             // (numeric) The nonce
	Bits              string          `json:"bits"`              // (string) The bits
	Difficulty        decimal.Decimal `json:"difficulty"`        // (numeric) The difficulty
	ChainWork         string          `json:"chainwork"`         // (string) Expected number of hashes required to produce the chain up to this block (hex)
	PreviousBlockHash string          `json:"previousblockhash"` // (string) The hash of the previous block (hex)
	NextBlockHash     string          `json:"nextblockhash"`     // (string) The hash of the next block (hex)
	Tx                []RawTxn        `json:"tx"`                // (json array) The transaction ids
}

type RawTxn struct {
	TxID     string       `json:"txid"`     // The transaction id
	Hash     string       `json:"hash"`     // The transaction hash (differs from txid for witness transactions)
	Size     int64        `json:"size"`     // The transaction size
	VSize    int64        `json:"vsize"`    // The virtual transaction size (differs from size for witness transactions)
	Version  int64        `json:"version"`  // The version
	LockTime int64        `json:"locktime"` // The lock time
	VIn      []RawTxnVIn  `json:"vin"`      // Array of transaction inputs (UTXOs to spend)
	VOut     []RawTxnVOut `json:"vout"`     // Array of transaction outputs (UTXOs to create)
}
type RawTxnVIn struct {
	TxID        string          `json:"txid"`        // The transaction id (UTXO)
	VOut        int             `json:"vout"`        // The output number (UTXO)
	ScriptSig   RawTxnScriptSig `json:"scriptSig"`   // The "signature script" (solution to the UTXO "pubkey script")
	TxInWitness []string        `json:"txinwitness"` // Array of hex-encoded witness data (if any)
	Sequence    int64           `json:"sequence"`    // The script sequence number
}
type RawTxnScriptSig struct {
	Asm string `json:"asm"` // The script disassembly
	Hex string `json:"hex"` // The script hex
}
type RawTxnVOut struct {
	Value        decimal.Decimal    `json:"value"`        // The value in DOGE (an exact decimal number)
	N            int                `json:"n"`            // The output number (VOut when spending)
	ScriptPubKey RawTxnScriptPubKey `json:"scriptPubKey"` // The "pubkey script" (conditions for spending this output)
}
type RawTxnScriptPubKey struct {
	Asm       string   `json:"asm"`       // The script disassembly
	Hex       string   `json:"hex"`       // The script hex
	ReqSigs   int64    `json:"reqSigs"`   // Number of required signatures
	Type      string   `json:"type"`      // Core RPC Script Type (see DecodeCoreRPCScriptType) NB. does NOT match our ScriptType enum!
	Addresses []string `json:"addresses"` // Array of dogecoin addresses accepted by the script
}

type BlockHeader struct {
	Hash              string          `json:"hash"`              // (string) the block hash (same as provided) (hex)
	Confirmations     int64           `json:"confirmations"`     // (numeric) The number of confirmations, or -1 if the block is not on the main chain
	Height            int64           `json:"height"`            // (numeric) The block height or index
	Version           int             `json:"version"`           // (numeric) The block version
	VersionHex        string          `json:"versionHex"`        // (string) The block version formatted in hexadecimal
	MerkleRoot        string          `json:"merkleroot"`        // (string) The merkle root (hex)
	Time              int             `json:"time"`              // (numeric) The block time in seconds since UNIX epoch (Jan 1 1970 GMT)
	MedianTime        int             `json:"mediantime"`        // (numeric) The median block time in seconds since UNIX epoch (Jan 1 1970 GMT)
	Nonce             int             `json:"nonce"`             // (numeric) The nonce
	Bits              string          `json:"bits"`              // (string) The bits
	Difficulty        decimal.Decimal `json:"difficulty"`        // (numeric) The difficulty
	ChainWork         string          `json:"chainwork"`         // (string) Expected number of hashes required to produce the chain up to this block (hex)
	PreviousBlockHash string          `json:"previousblockhash"` // (string) The hash of the previous block (hex)
	NextBlockHash     string          `json:"nextblockhash"`     // (string) The hash of the next block (hex)
}

func (b *BlockHeader) IsOnChain() bool {
	return b.Confirmations != -1
}

type UTXO struct {
	TxID          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Amount        float64 `json:"amount"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	RedeemScript  string  `json:"redeemScript,omitempty"`
	Spendable     bool    `json:"spendable"`
	Solvable      bool    `json:"solvable"`
	Desc          string  `json:"desc"`
	Safe          bool    `json:"safe"`
	Confirmations int     `json:"confirmations"`
}

type Info struct {
	Version         int64   `json:"version"`
	ProtocolVersion int64   `json:"protocolversion"`
	WalletVersion   int64   `json:"walletversion"`
	Balance         float64 `json:"balance"`
	Blocks          int64   `json:"blocks"`
	TimeOffset      int64   `json:"timeoffset"`
	Connections     int64   `json:"connections"`
	Proxy           string  `json:"proxy"`
	Difficulty      float64 `json:"difficulty"`
	Testnet         bool    `json:"testnet"`
	KeypoolOldest   int64   `json:"keypoololdest"`
	KeypoolSize     int64   `json:"keypoolsize"`
	PayTxFee        float64 `json:"paytxfee"`
	RelayFee        float64 `json:"relayfee"`
	Errors          string  `json:"errors"`
}

type BlockFilter struct {
	Filter string `json:"filter"`
	Header string `json:"header"`
}

type BlockStats struct {
	AvgFee             float64   `json:"avgfee"`
	AvgFeerate         float64   `json:"avgfeerate"`
	Avgtxsize          float64   `json:"avgtxsize"`
	Blockhash          string    `json:"blockhash"`
	FeeratePercentiles []float64 `json:"feerate_percentiles"`
	Height             int64     `json:"height"`
	Ins                int64     `json:"ins"`
	Maxfee             float64   `json:"maxfee"`
	Maxfeerate         float64   `json:"maxfeerate"`
	Maxtxsize          float64   `json:"maxtxsize"`
	Medianfee          float64   `json:"medianfee"`
	Mediantxsize       float64   `json:"mediantxsize"`
	Minfee             float64   `json:"minfee"`
	Minfeerate         float64   `json:"minfeerate"`
	Mintxsize          float64   `json:"mintxsize"`
	Outs               int64     `json:"outs"`
	Subsidy            float64   `json:"subsidy"`
	SwtotalSize        float64   `json:"swtotal_size"`
	SwtotalWeight      float64   `json:"swtotal_weight"`
	Swtxs              int64     `json:"swtxs"`
	Time               int64     `json:"time"`
	TotalOut           float64   `json:"total_out"`
	TotalSize          float64   `json:"total_size"`
	TotalWeight        float64   `json:"total_weight"`
	Totalfee           float64   `json:"totalfee"`
	Txs                int64     `json:"txs"`
	UtxoIncrease       int64     `json:"utxo_increase"`
	UtxoSizeInc        int64     `json:"utxo_size_inc"`
}

type ChainTip struct {
	Height    int64  `json:"height"`
	Hash      string `json:"hash"`
	BranchLen int    `json:"branchlen"`
	Status    string `json:"status"`
}

type ChainTxStats struct {
	Time                   int64  `json:"time"`
	TxCount                int64  `json:"txcount"`
	WindowFinalBlockHash   string `json:"window_final_block_hash"`
	WindowFinalBlockHeight int64  `json:"window_final_block_height"`
	WindowBlockCount       int64  `json:"window_block_count"`
	WindowTxCount          int64  `json:"window_tx_count"`
	WindowInterval         int64  `json:"window_interval"`
	TxRate                 int64  `json:"txrate"`
}

type MempoolAncestors struct {
	TxID            string `json:"transactionid"`
	VSize           int64  `json:"vsize"`
	Weight          int64  `json:"weight"`
	Fee             int64  `json:"fee"`
	ModifiedFee     int64  `json:"modifiedfee"`
	Time            int64  `json:"time"`
	Height          int64  `json:"height"`
	DescendantCount int64  `json:"descendantcount"`
	DescendantSize  int64  `json:"descendantsize"`
	DescendantFees  int64  `json:"descendantfees"`
	AncestorCount   int64  `json:"ancestorcount"`
	AncestorSize    int64  `json:"ancestorsize"`
	AncestorFees    int64  `json:"ancestorfees"`
	Wtxid           string `json:"wtxid"`
	Fees            struct {
		Base       int64 `json:"base"`
		Modified   int64 `json:"modified"`
		Ancestor   int64 `json:"ancestor"`
		Descendant int64 `json:"descendant"`
	} `json:"fees"`
	Depends           []string `json:"depends"`
	Spentby           []string `json:"spentby"`
	Bip125Replaceable bool     `json:"bip125-replaceable"`
	Unbroadcast       bool     `json:"unbroadcast"`
}

type MempoolInfo struct {
	Loaded           bool    `json:"loaded"`
	Size             int64   `json:"size"`
	Bytes            int64   `json:"bytes"`
	Usage            int64   `json:"usage"`
	MaxMempool       int64   `json:"maxmempool"`
	MempoolMinFee    float32 `json:"mempoolminfee"`
	MinRelayTxfee    int64   `json:"minrelaytxfee"`
	UnbroadcastCount int64   `json:"unbroadcastcount"`
}

type RawMempoolInfo struct {
	TransactionID map[string]struct {
		VSize       int64 `json:"vsize"`
		Weight      int64 `json:"weight"`
		Fee         int64 `json:"fee"`
		ModifiedFee int64 `json:"modifiedfee"`
		Time        int64 `json:"time"`
	} `json:"transactionid"`
	Depends           []string `json:"depends"`
	Spentby           []string `json:"spentby"`
	Bip125Replaceable bool     `json:"bip125-replaceable"`
	Unbroadcast       bool     `json:"unbroadcast"`
}

type TxOut struct {
	BestBlock     string             `json:"bestblock"`
	Confirmations int64              `json:"confirmations"`
	Value         int64              `json:"value"`
	ScriptPubKey  RawTxnScriptPubKey `json:"scriptPubKey"`
	Coinbase      bool               `json:"coinbase"`
}

type WalletInfo struct {
	WalletName string          `json:"walletname"`
	Balance    decimal.Decimal `json:"balance"`
	TxCount    int64           `json:"txcount"`
}

type FundRawTransactionResponse struct {
	Hex       string  `json:"hex"`
	Fee       float64 `json:"fee"`
	ChangePos float64 `json:"changepos"`
}

type SignRawTransactionResponse struct {
	Hex string `json:"hex"`
}
