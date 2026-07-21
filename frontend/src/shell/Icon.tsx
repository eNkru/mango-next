import type { LucideIcon } from 'lucide-react';

export type IconProps = {
  icon: LucideIcon;
  size?: number;
  className?: string;
  /** Default true → aria-hidden. Set false or pass `label` for accessible icons. */
  decorative?: boolean;
  /** When set, icon is exposed as role=img with aria-label (decorative ignored). */
  label?: string;
};

export function Icon({
  icon: Lucide,
  size = 18,
  className,
  decorative = true,
  label,
}: IconProps) {
  // Always pass a boolean so Lucide defaults cannot leave aria-hidden=true
  // when the icon is meant to be accessible via `label`.
  const isDecorative = label ? false : decorative;
  return (
    <Lucide
      size={size}
      className={className ? `mango-icon ${className}` : 'mango-icon'}
      aria-hidden={isDecorative}
      role={label ? 'img' : undefined}
      aria-label={label}
      focusable="false"
    />
  );
}
