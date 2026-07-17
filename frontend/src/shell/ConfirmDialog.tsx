type ConfirmDialogProps = {
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
};

export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = '确认',
  cancelLabel = '取消',
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  if (!open) return null;

  return (
    <div className="mango-modal-backdrop" role="presentation" onClick={onCancel}>
      <div
        className="mango-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="mango-confirm-title"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="mango-confirm-title">{title}</h2>
        <p>{message}</p>
        <div className="mango-modal__actions">
          <button type="button" className="mango-btn" onClick={onCancel}>
            {cancelLabel}
          </button>
          <button type="button" className="mango-btn mango-btn--danger" onClick={onConfirm}>
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
