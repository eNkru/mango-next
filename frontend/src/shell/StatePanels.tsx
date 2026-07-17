type Props = {
  message?: string;
};

export function LoadingState({ message = '加载中…' }: Props) {
  return <div className="mango-state">{message}</div>;
}

export function EmptyState({ message = '暂无数据' }: Props) {
  return <div className="mango-state">{message}</div>;
}

export function ErrorState({ message = '出错了' }: Props) {
  return (
    <div className="mango-state mango-state--error" role="alert">
      {message}
    </div>
  );
}
