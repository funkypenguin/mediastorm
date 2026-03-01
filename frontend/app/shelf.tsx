import MobileTabBar from '@/components/MobileTabBar';
import WatchlistScreen from './(drawer)/watchlist';
import { Platform, Pressable } from 'react-native';

export default function ShelfScreen() {
  return (
    <>
      {/* Android TV focus anchor — shelf is outside the drawer layout,
          so it needs its own native focusable element for D-pad events to flow */}
      {Platform.OS === 'android' && Platform.isTV && (
        <Pressable
          style={{ position: 'absolute', width: 1, height: 1, opacity: 0 }}
          accessible={false}
          importantForAccessibility="no"
          focusable={true}
        />
      )}
      <WatchlistScreen />
      <MobileTabBar />
    </>
  );
}
