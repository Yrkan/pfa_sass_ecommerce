import { useState } from "react";
import Panel from "./Panel";

function App() {
  const [id, setId] = useState("")
  const [ready, setReady] = useState(true)
  return (
    <div className="App">
      <header className="App-header">
          <input type="text" value={id} onChange={(e) => setId(e.target.value)}/>
          <button onClick={() => setReady(true)}>ok</button>
      </header>
      
      {ready && <section>
        <Panel link={id} />
        </section>}
    </div>
  );
}

export default App;
