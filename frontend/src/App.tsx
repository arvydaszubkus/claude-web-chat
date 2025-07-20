import { useState } from "react";
import "./App.css";

interface Message {
  role: string;
  content: string;
}

function App() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");

  const sendMessage = async () => {
    const newMessages = [...messages, { role: "user", content: input }];
    setInput("");
    setMessages(newMessages);

    const res = await fetch("http://localhost:8080/api/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ messages: newMessages }),
    });

    const data = await res.json();
    setMessages([...newMessages, { role: "assistant", content: data.reply }]);
  };

  return (
    <div className="App">
      <h2>Claude Web Chat</h2>
      <div className="chat-window">
        {messages.map((m, i) => (
          <div key={i} className={m.role}>
            <strong>{m.role}:</strong> {m.content}
          </div>
        ))}
      </div>
      <div className="input-area">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && sendMessage()}
          placeholder="Ask Claude..."
        />
        <button onClick={sendMessage}>Send</button>
      </div>
    </div>
  );
}


export default App;
