type LoadingProps = {
  message?: string;
};

type EmptyProps = {
  message?: string;
};

type ErrorProps = {
  message?: string;
  onRetry?: () => void;
  retryLabel?: string;
};

export function LoadingState({ message = '…' }: LoadingProps) {
  return <div className="mango-state">{message}</div>;
}

export function EmptyState({ message = '…' }: EmptyProps) {
  return <div className="mango-state">{message}</div>;
}

export function ErrorState({ message = '…', onRetry, retryLabel = 'Retry' }: ErrorProps) {
  return (
    <div className="mango-state mango-state--error" role="alert">
      <div>{message}</div>
      {onRetry ? (
        <button type="button" className="mango-btn" onClick={onRetry}>
          {retryLabel}
        </button>
      ) : null}
    </div>
  );
}
