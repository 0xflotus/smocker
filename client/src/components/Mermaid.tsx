import { Alert } from "antd";
import mermaidAPI from "mermaid";
import * as React from "react";

export const Mermaid = ({ name, chart }: any) => {
  const [diagram, setDiagram] = React.useState("");
  const [error, setError] = React.useState("");

  React.useEffect(() => {
    const cb = (svg?: string) => {
      setDiagram(svg || "");
      setError("");
    };
    try {
      mermaidAPI.parse(chart);
      mermaidAPI.initialize({ startOnLoad: false });
      mermaidAPI.render(name, chart, cb);
    } catch (e) {
      setDiagram("");
      console.error(e);
      setError(e.str || `${e}`);
    }
  }, [name, chart]);

  return (
    <div className="mermaid">
      <div
        style={{ width: "100%" }}
        dangerouslySetInnerHTML={{ __html: diagram }}
      />
      {error && (
        <Alert
          message="Unable to render"
          description={error}
          type="error"
          showIcon
        />
      )}
    </div>
  );
};