export type ReaderDimension = {
  width: number;
  height: number;
};

export type ReaderEntry = {
  id: string;
  name: string;
  pages: number;
  progress: number;
};

export type ReaderBootstrap = {
  title: { id: string; name: string };
  entry: ReaderEntry;
  dimensions: ReaderDimension[];
  entries: ReaderEntry[];
  exit_url: string;
  next_entry_url: string;
  previous_entry_url: string;
};

export type ReaderMode = 'continuous' | 'paged';
export type ReaderFitType = 'vert' | 'horz' | 'original';

export type ReaderPrefs = {
  mode: ReaderMode;
  margin: number;
  fitType: ReaderFitType;
  preloadLookahead: number;
  enableFlipAnimation: boolean;
  enableRightToLeft: boolean;
};
