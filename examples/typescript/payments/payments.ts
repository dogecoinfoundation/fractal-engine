import km2, {
    Crypto,
    Net,
    Language,
    UnsignedTransaction,
    Wallet,
} from "@houseofdoge/km2";
import { canonicalize } from "json-canonicalize";

const FE_URL = "http://HOST_OF_YOUR_FRACTAL_ENGINE:PORT_OF_YOUR_FRACTAL_ENGINE";
const KOINU = 100_000_000;

async function main() {
    using seed = new km2.SeedPhrase({
        wordCount: 24,
        language: Language.English,
    });

    using wallet = new km2.Wallet(seed, {
        cryptocurrency: Crypto.Dogecoin,
        network: Net.Mainnet
    });

    using kp = wallet.deriveKeypair({ account: 1, change: 0, index: 0 });

    // Example UTXO, you will need to retrieve this from your wallet
    const utxos = [
        {
            value: "1.0",
            vout: 0,
            tx: "0x0000000000000000000000000000000000000000000000000000000000000000",
            script: "0x0000000000000000000000000000000000000000000000000000000000000000",
        }
    ];

    const invoiceHash = "INVOICE_HASH";
    const invoiceTotal = 25;
    const sellerAddress = "SELLER_ADDRESS";

    const paymentResponse = await payInvoiceHttp(invoiceHash);
    const unsignedTrxn = new UnsignedTransaction(Crypto.Dogecoin, Net.Mainnet);

    // NOTE: This is a pretty crude way of figuring out UTXOs and Fees.
    const totalValue = dogeToKoinu(utxos[0].value);
    const totalFee = dogeToKoinu("0.002");
    const invoiceValue = dogeToKoinu(`${invoiceTotal}`);

    const changeValue = totalValue - invoiceValue - totalFee;

    unsignedTrxn.addInput({
        outputIndex: utxos[0].vout,
        prevTxId: utxos[0].tx,
        scriptPubKeyHex: utxos[0].script,
        value: totalValue,
        sequence: 0xffffffff,
    });

    unsignedTrxn.addOutput({
        kind: "payment",
        address: kp.address,
        value: changeValue,
    });

    unsignedTrxn.addOutput({
        kind: "payment",
        address: sellerAddress,
        value: invoiceValue,
    });

    unsignedTrxn.addOutput({
        kind: "opReturn",
        data: paymentResponse.encoded_transaction_body,
        value: 0,
    });

    const signedTrxn = unsignedTrxn.sign({
        keypairs: [kp],
    });

    const trxnId = await sendSignedTransaction(signedTrxn.rawHex);


    console.log(trxnId);
}


const dogeToKoinu = (s: string): number => {
    const m = s.trim().match(/^([+-])?(\d+)(?:\.(\d{0,8}))?$/);
    if (!m) throw new Error(`invalid amount: ${s}`);
    const sign = m[1] === "-" ? -1 : 1;
    const w = m[2].replace(/^0+/, "") || "0";
    const f = (m[3] || "").padEnd(8, "0");

    // Max safe: whole <= 90,071,992 (because (whole*1e8 + frac) must fit 2^53-1)
    if (w.length > 8 || (w.length === 8 && Number(w) > 90071992)) {
        throw new Error("amount too large for a JS number; use the string version");
    }

    return sign * (Number(w) * KOINU + Number(f));
};

const koinuToDoge = (k: number): string =>
    `${k / KOINU}.${(k % KOINU).toString().padStart(8, "0")}`;


const payInvoiceHttp = async (
    invoiceHash: string,
): Promise<{ encoded_transaction_body: string }> => {
    const feUrl = FE_URL;

    let invoiceEnvelope = {
        invoice_hash: invoiceHash,
    };

    const res = await fetch(feUrl + "/payments/new", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Accept: "application/json",
        },
        body: JSON.stringify(invoiceEnvelope),
    });

    const resJson = await res.json();

    return resJson;
};


const sendSignedTransaction = async (
    encodedTrxnHex: string,
): Promise<string> => {

    const res = await fetch(FE_URL + "/doge/send", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Accept: "application/json",
        },
        body: JSON.stringify({
            encoded_transaction_hex: encodedTrxnHex,
        }),
    });

    const resJson = await res.json();

    return resJson.raw_transaction_hex;
};

export async function sha256Hash(input: string | Uint8Array): Promise<string> {
    const buf = typeof input === "string" ? new TextEncoder().encode(input) : input;
    const hashBuffer = await crypto.subtle.digest("SHA-256", buf as BufferSource);
    const hashArray = new Uint8Array(hashBuffer);
    return Array.from(hashArray).map(b => b.toString(16).padStart(2, '0')).join('');
}

export function jsonStringifyCanonical(payload: unknown): Uint8Array {
    const canon = canonicalize(payload); // RFC 8785 canonical JSON
    return new TextEncoder().encode(canon); // UTF-8
}

main().then(() => {
    console.log("done");
}).catch((err) => {
    console.error(err);
});