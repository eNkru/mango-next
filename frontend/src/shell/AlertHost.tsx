import { useEffect, useState } from 'react';

export type AlertTone = 'info' | 'success' | 'danger';

export type AlertItem = {
  id: number;
  tone: AlertTone;
  message: string;
};

type Listener = (item: AlertItem) => void;

const listeners = new Set<Listener>();
let nextId = 1;

export function pushAlert(message: string, tone: AlertTone = 'info'): void {
  const item = { id: nextId++, tone, message };
  listeners.forEach((fn) => fn(item));
}

export function AlertHost() {
  const [items, setItems] = useState<AlertItem[]>([]);

  useEffect(() => {
    const onAlert: Listener = (item) => {
      setItems((prev) => [...prev, item]);
      window.setTimeout(() => {
        setItems((prev) => prev.filter((x) => x.id !== item.id));
      }, 4000);
    };
    listeners.add(onAlert);
    return () => {
      listeners.delete(onAlert);
    };
  }, []);

  if (!items.length) return null;

  return (
    <div className="mango-alert-stack" role="status" aria-live="polite">
      {items.map((item) => (
        <div key={item.id} className={`mango-alert mango-alert--${item.tone}`}>
          {item.message}
        </div>
      ))}
    </div>
  );
}
