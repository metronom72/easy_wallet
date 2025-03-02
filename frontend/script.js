const BOT_TOKEN = "";

function parseQueryParams() {
    console.log("Parsing query parameters...");
    const params = new URLSearchParams(window.location.search);
    let authData = {};
    for (const [key, value] of params.entries()) {
        authData[key] = value;
    }
    console.log("Parsed Authentication Data:", authData);
    return authData;
}

async function sha256(input) {
    console.log("Computing SHA-256 hash of bot token...");
    const encoder = new TextEncoder();
    const data = encoder.encode(input);
    const hashBuffer = await crypto.subtle.digest("SHA-256", data);
    console.log("SHA-256 Hash:", new Uint8Array(hashBuffer));
    return hashBuffer;
}

async function hmacSha256(key, message) {
    console.log("Computing HMAC-SHA256...");
    console.log("Message for HMAC:", message);

    const encoder = new TextEncoder();
    const cryptoKey = await crypto.subtle.importKey(
        "raw",
        key,
        { name: "HMAC", hash: "SHA-256" },
        false,
        ["sign"]
    );

    const signature = await crypto.subtle.sign(
        "HMAC",
        cryptoKey,
        encoder.encode(message)
    );

    const hexSignature = [...new Uint8Array(signature)]
        .map((b) => b.toString(16).padStart(2, "0"))
        .join("");

    console.log("Computed HMAC-SHA256:", hexSignature);
    return hexSignature;
}

function constructDataCheckString(authData) {
    console.log("Constructing data_check_string...");
    const { hash, ...dataWithoutHash } = authData;
    const sortedData = Object.entries(dataWithoutHash)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([key, value]) => `${key}=${value}`)
        .join("\n");

    console.log("Constructed data_check_string:\n", sortedData);
    return sortedData;
}

async function verifyTelegramData(authData) {
    if (!authData.hash) {
        console.error("Hash is missing in authentication data.");
        return false;
    }

    const receivedHash = authData.hash.toLowerCase();
    console.log("Received Hash from URL:", receivedHash);

    const dataCheckString = constructDataCheckString(authData);
    const secretKey = await sha256(BOT_TOKEN);
    const expectedHash = await hmacSha256(secretKey, dataCheckString);

    console.log("Expected Hash:", expectedHash);
    console.log("Comparing hashes...");

    if (expectedHash === receivedHash) {
        console.log("✅ Verification successful!");
        return true;
    } else {
        console.warn("❌ Verification failed!");
        return false;
    }
}

window.onload = async function () {
    console.log("Window loaded. Initializing authentication check...");
    const authData = parseQueryParams();

    const sortedConcatenatedData = Object.entries(authData)
        .sort(([a], [b]) => a.localeCompare(b))
        .filter(([key, value]) => key !== 'hash')
        .map(([key, value]) => `${key}=${value}`)
        .join("\\n");

    document.getElementById("auth-data").textContent = JSON.stringify(authData, null, 4)

    document.getElementById("sorted-auth-data").textContent = sortedConcatenatedData;

    document.getElementById("verify-btn").addEventListener("click", async () => {
        console.log("Verify button clicked.");
        const resultElement = document.getElementById("verification-result");

        try {
            const isValid = await verifyTelegramData(authData);
            resultElement.textContent = isValid
                ? "✅ Data is verified!"
                : "❌ Data verification failed!";
        } catch (error) {
            console.error("Error during verification:", error);
            resultElement.textContent = "⚠️ Error verifying data!";
        }
    });
};
