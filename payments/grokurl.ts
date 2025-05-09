async function askGrokX(question: string) {
    const apiKey = "YOUR_X_API_KEY";
    const url = "https://api.x.ai/v2/grok/chat";

    const requestBody = {
        prompt: question,
        user_id: "your_user_id",
    };

    try {
        const response = await fetch(url, {
            method: "POST",
            headers: {
                "Authorization": `Bearer ${apiKey}`,
                "Content-Type": "application/json",
            },
            body: JSON.stringify(requestBody),
        });

        if (!response.ok) {
            throw new Error(`HTTP Error: ${response.status}`);
        }

        const data = await response.json();
        console.log("Response:", data);

        return data?.message || "No response received.";
    } catch (error) {
        console.error("Error fetching Grok response:", error);
        return "Error occurred.";
    }
}

askGrokX("How much is 2 + 2 ?").then(console.log);
