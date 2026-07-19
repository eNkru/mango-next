import {
  apiIndexToUrlPage,
  clampPage,
  isLongPageTitle,
  nextDirectionIsLeft,
  shouldSaveProgress,
  urlPageToApiIndex,
} from './readerMath';

function assert(cond: unknown, msg: string): asserts cond {
  if (!cond) throw new Error(msg);
}

function run() {
  assert(clampPage(0, 10) === 1, 'clamp low');
  assert(clampPage(99, 10) === 10, 'clamp high');
  assert(clampPage(5, 10) === 5, 'clamp mid');
  assert(clampPage(NaN, 10) === 1, 'clamp nan');

  assert(urlPageToApiIndex(3) === 3, 'url->api');
  assert(apiIndexToUrlPage(3) === 3, 'api->url');

  assert(nextDirectionIsLeft(false) === false, 'ltr');
  assert(nextDirectionIsLeft(true) === true, 'rtl');

  assert(shouldSaveProgress(1, 10, 100, false) === true, 'first');
  assert(shouldSaveProgress(100, 10, 100, false) === true, 'last');
  assert(shouldSaveProgress(15, 10, 100, false) === true, 'distance');
  assert(shouldSaveProgress(12, 10, 100, false) === false, 'throttle');
  assert(shouldSaveProgress(12, 10, 100, true) === true, 'long pages');

  assert(isLongPageTitle([{ width: 100, height: 300 }, { width: 100, height: 250 }]) === true, 'long');
  assert(isLongPageTitle([{ width: 1000, height: 1500 }]) === false, 'normal');

  console.log('readerMath tests passed');
}

run();
