import km2, {
    Crypto,
    Net,
    Language,
} from "@houseofdoge/km2";
import { canonicalize } from "json-canonicalize";
 

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

// This will generate a new wallet (purely for example purposes)
using seed = new km2.SeedPhrase({
    wordCount: 24,
    language: Language.English,
});

using wallet = new km2.Wallet(seed, {
    cryptocurrency: Crypto.Dogecoin,
    network: Net.Mainnet
});

using kp = wallet.deriveKeypair({ account: 1, change: 0, index: 0 });

const myBodyData = {
    name: "Test",
    description: "Test",
}

const hashedPayload = await sha256Hash(jsonStringifyCanonical(myBodyData));

const signature = kp.signMessage({
    message: hashedPayload,
});

console.log(signature.toBase64());