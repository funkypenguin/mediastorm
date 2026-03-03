import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { Animated, BackHandler, Dimensions, Platform, StyleSheet, View } from 'react-native';
import { StatusBar } from 'expo-status-bar';
import { Gesture, GestureDetector } from 'react-native-gesture-handler';
import StrmrLoadingScreen from '@/app/strmr-loading';
import { useBackendSettings } from './BackendSettingsContext';

const SCREEN_WIDTH = Dimensions.get('window').width;
const FADE_IN_DURATION = 300;
const FADE_OUT_DURATION = 400;

type LoadingScreenContextValue = {
  showLoadingScreen: () => void;
  hideLoadingScreen: (options?: { immediate?: boolean }) => void;
  isLoadingScreenVisible: boolean;
  setOnCancel: (callback: (() => void) | null) => void;
};

const LoadingScreenContext = createContext<LoadingScreenContextValue | null>(null);

export function useLoadingScreen(): LoadingScreenContextValue {
  const context = useContext(LoadingScreenContext);
  if (!context) {
    throw new Error('useLoadingScreen must be used within a LoadingScreenProvider.');
  }
  return context;
}

type LoadingScreenProviderProps = {
  children: ReactNode;
};

export function LoadingScreenProvider({ children }: LoadingScreenProviderProps) {
  // isVisible = logical state (true while loading screen should be shown)
  // isRendered = keeps overlay mounted during fade-out animation
  const [isVisible, setIsVisible] = useState(false);
  const [isRendered, setIsRendered] = useState(false);
  const [onCancelCallback, setOnCancelCallback] = useState<(() => void) | null>(null);
  const { settings, userSettings } = useBackendSettings();
  const isLoadingScreenEnabled =
    userSettings?.playback?.useLoadingScreen ?? settings?.playback?.useLoadingScreen ?? false;
  const translateX = useRef(new Animated.Value(0)).current;
  const opacity = useRef(new Animated.Value(0)).current;

  // Track desired visibility — the effect below drives the actual animations
  const wantsVisibleRef = useRef(false);
  const fadeOutTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const showLoadingScreen = useCallback(() => {
    if (isLoadingScreenEnabled) {
      // Cancel any in-progress fade-out
      if (fadeOutTimerRef.current) {
        clearTimeout(fadeOutTimerRef.current);
        fadeOutTimerRef.current = null;
      }
      wantsVisibleRef.current = true;
      translateX.setValue(0);
      opacity.setValue(0);
      setIsVisible(true);
      setIsRendered(true);
    }
  }, [isLoadingScreenEnabled, translateX, opacity]);

  // Fade in once the overlay is mounted
  useEffect(() => {
    if (isRendered && wantsVisibleRef.current) {
      requestAnimationFrame(() => {
        Animated.timing(opacity, {
          toValue: 1,
          duration: FADE_IN_DURATION,
          useNativeDriver: true,
        }).start();
      });
    }
  }, [isRendered, opacity]);

  const hideLoadingScreen = useCallback((options?: { immediate?: boolean }) => {
    wantsVisibleRef.current = false;
    setIsVisible(false);
    if (options?.immediate) {
      opacity.stopAnimation();
      opacity.setValue(0);
      setIsRendered(false);
      translateX.setValue(0);
    } else {
      Animated.timing(opacity, {
        toValue: 0,
        duration: FADE_OUT_DURATION,
        useNativeDriver: true,
      }).start(() => {
        if (!wantsVisibleRef.current) {
          setIsRendered(false);
          translateX.setValue(0);
        }
      });
    }
  }, [translateX, opacity]);

  const setOnCancel = useCallback((callback: (() => void) | null) => {
    setOnCancelCallback(() => callback);
  }, []);

  const handleCancel = useCallback(() => {
    // Slide out then hide
    Animated.timing(translateX, {
      toValue: SCREEN_WIDTH,
      duration: 250,
      useNativeDriver: true,
    }).start(() => {
      if (onCancelCallback) {
        onCancelCallback();
      }
      setIsVisible(false);
      setIsRendered(false);
      translateX.setValue(0);
      opacity.setValue(0);
    });
  }, [onCancelCallback, translateX, opacity]);

  // Handle Android back button when loading screen is visible
  useEffect(() => {
    if (!isVisible) return;
    const subscription = BackHandler.addEventListener('hardwareBackPress', () => {
      handleCancel();
      return true;
    });
    return () => subscription.remove();
  }, [isVisible, handleCancel]);

  const swipeGesture = useMemo(
    () =>
      Gesture.Pan()
        .activeOffsetX(10)
        .onChange((event) => {
          if (event.translationX > 0) {
            translateX.setValue(event.translationX);
          }
        })
        .onEnd((event) => {
          const shouldDismiss = event.translationX > SCREEN_WIDTH * 0.3 || event.velocityX > 500;

          if (shouldDismiss) {
            Animated.timing(translateX, {
              toValue: SCREEN_WIDTH,
              duration: 200,
              useNativeDriver: true,
            }).start(() => {
              if (onCancelCallback) {
                onCancelCallback();
              }
              setIsVisible(false);
              setIsRendered(false);
              translateX.setValue(0);
              opacity.setValue(0);
            });
          } else {
            Animated.spring(translateX, {
              toValue: 0,
              useNativeDriver: true,
              damping: 20,
              stiffness: 300,
            }).start();
          }
        })
        .runOnJS(true),
    [translateX, opacity, onCancelCallback],
  );

  const value = useMemo<LoadingScreenContextValue>(
    () => ({
      showLoadingScreen,
      hideLoadingScreen,
      isLoadingScreenVisible: isVisible,
      setOnCancel,
    }),
    [hideLoadingScreen, showLoadingScreen, isVisible, setOnCancel],
  );

  return (
    <LoadingScreenContext.Provider value={value}>
      <View style={styles.rootContainer}>
        {children}
        {isRendered && (
          <GestureDetector gesture={swipeGesture}>
            <Animated.View
              renderToHardwareTextureAndroid={true}
              style={[
                styles.overlay,
                {
                  opacity,
                  transform: [{ translateX }],
                },
              ]}>
              <StatusBar hidden />
              <StrmrLoadingScreen embedded />
            </Animated.View>
          </GestureDetector>
        )}
      </View>
    </LoadingScreenContext.Provider>
  );
}

const styles = StyleSheet.create({
  rootContainer: {
    flex: 1,
  },
  overlay: {
    ...StyleSheet.absoluteFillObject,
    zIndex: 9999,
    backgroundColor: '#000000',
  },
});
