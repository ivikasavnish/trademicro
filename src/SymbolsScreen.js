import React, { useEffect, useState } from 'react';
import { View, Text, FlatList, StyleSheet, ActivityIndicator } from 'react-native';
import api from './api';

export default function SymbolsScreen() {
  const [symbols, setSymbols] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api.get('/symbols')
      .then(response => setSymbols(response.data))
      .catch(err => setError('Failed to fetch symbols'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <ActivityIndicator />;
  if (error) return <Text>{error}</Text>;

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Symbols</Text>
      <FlatList
        data={symbols}
        keyExtractor={item => item}
        renderItem={({ item }) => <Text style={styles.item}>{item}</Text>}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, padding: 24 },
  title: { fontSize: 24, marginBottom: 12 },
  item: { fontSize: 18, padding: 8 },
});
