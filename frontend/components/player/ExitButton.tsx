import React, { useMemo } from 'react';
import { Platform, Pressable, StyleSheet, Text } from 'react-native';
import FocusablePressable from '@/components/FocusablePressable';
import { useTheme } from '@/theme';
import type { NovaTheme } from '@/theme';
import { useTVDimensions } from '@/hooks/useTVDimensions';
import { TV_REFERENCE_HEIGHT } from '@/theme/tokens/tvScale';

interface GoBackButtonProps {
  onSelect: () => void;
  onFocus?: () => void;
  disabled?: boolean;
}

const ExitButton: React.FC<GoBackButtonProps> = ({ onSelect, onFocus, disabled }) => {
  const theme = useTheme();
  const { height } = useTVDimensions();
  const vh = (height > 0 ? height : 1080) / TV_REFERENCE_HEIGHT;
  const styles = useMemo(() => useExitButtonStyles(theme, vh), [theme, vh]);

  // Use FocusablePressable for all platforms.
  // It now handles TV platforms (tvOS and Android TV) via SpatialNavigationFocusableView internally.
  return (
    <FocusablePressable
      text={'Exit'}
      icon="arrow-back"
      focusKey="exit-button"
      onSelect={onSelect}
      onFocus={onFocus}
      disabled={disabled}
      style={styles.exitBtn}
      textStyle={styles.exitText}
      focusedTextStyle={styles.exitText}
    />
  );
};

const useExitButtonStyles = (theme: NovaTheme, vh: number) => {
  const isTVPlatform = Platform.isTV;
  return StyleSheet.create({
    exitBtn: {
      position: 'absolute',
      top: isTVPlatform ? Math.round(theme.spacing.lg * vh) : theme.spacing.lg,
      left: isTVPlatform ? Math.round(theme.spacing.lg * vh) : theme.spacing.lg,
      paddingVertical: isTVPlatform ? Math.round(theme.spacing.md * vh) : theme.spacing.md,
      paddingHorizontal: isTVPlatform ? Math.round(theme.spacing.lg * vh) : theme.spacing.lg,
      marginHorizontal: isTVPlatform ? Math.round(theme.spacing.lg * vh) : 0,
    },
    exitText: {
      fontSize: isTVPlatform ? Math.round(16 * vh) : 16,
      lineHeight: isTVPlatform ? Math.round(21 * vh) : 21,
    },
  });
};

export default ExitButton;
