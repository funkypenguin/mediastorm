import { requireNativeModule, Platform, EventEmitter, type EventSubscription } from 'expo-modules-core';

interface ApkUpdaterModule {
  downloadAndInstall(url: string): Promise<void>;
  cleanDownloads(): void;
}

interface DownloadProgressEvent {
  percent: number;
  bytesDownloaded: number;
  totalBytes: number;
}

// Safely load the native module - only available on Android with native build
function loadNativeModule(): ApkUpdaterModule | null {
  if (Platform.OS !== 'android') {
    return null;
  }
  try {
    return requireNativeModule('ApkUpdater');
  } catch (e) {
    console.log('[ApkUpdater] Failed to load native module:', e);
    return null;
  }
}

const ApkUpdaterNative = loadNativeModule();

const emitter = ApkUpdaterNative
  ? new EventEmitter<{ onDownloadProgress: (event: DownloadProgressEvent) => void }>(ApkUpdaterNative as any)
  : null;

/**
 * Download an APK from the given URL and launch the system package installer.
 * Emits onDownloadProgress events during download.
 */
export async function downloadAndInstall(url: string): Promise<void> {
  if (!ApkUpdaterNative) {
    throw new Error('ApkUpdater is not available on this platform');
  }
  return ApkUpdaterNative.downloadAndInstall(url);
}

/**
 * Clean cached APK downloads.
 */
export function cleanDownloads(): void {
  if (ApkUpdaterNative) {
    ApkUpdaterNative.cleanDownloads();
  }
}

/**
 * Subscribe to download progress events.
 * Callback receives { percent, bytesDownloaded, totalBytes }.
 * Returns a Subscription that should be removed on cleanup.
 */
export function addProgressListener(callback: (event: DownloadProgressEvent) => void): EventSubscription | null {
  if (!emitter) return null;
  return emitter.addListener('onDownloadProgress', callback);
}

export default {
  downloadAndInstall,
  cleanDownloads,
  addProgressListener,
};
