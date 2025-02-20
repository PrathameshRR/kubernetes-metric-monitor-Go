'use client';

import { useState, useEffect } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';

interface NodeMetrics {
  metadata: {
    name: string;
  };
  usage: {
    cpu: string;
    memory: string;
  };
}

interface PodMetrics {
  metadata: {
    name: string;
  };
  usage: {
    cpu: string;
    memory: string;
  };
}

interface Metrics {
  nodes: NodeMetrics[];
  pods: PodMetrics[];
}

export default function Home() {
  const [metrics, setMetrics] = useState<Metrics>({ nodes: [], pods: [] });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        const basePath = process.env.NODE_ENV === 'production' 
          ? '' 
          : 'http://localhost:8080';
    
        const response = await fetch(`${basePath}/api/metrics`);
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        
        // Validate the data structure
        if (!data || !Array.isArray(data.nodes) || !Array.isArray(data.pods)) {
          throw new Error('Invalid metrics data format');
        }
        
        setMetrics(data);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setLoading(false);
      }
    };

    fetchMetrics();
    const interval = setInterval(fetchMetrics, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl">Loading metrics...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-xl text-red-500">Error: {error}</div>
      </div>
    );
  }

  const nodeData = metrics.nodes.map(node => {
    // Add null checks and default values
    const cpuValue = node?.usage?.cpu || '0m';
    const memoryValue = node?.usage?.memory || '0Mi';
    
    return {
      name: node?.metadata?.name || 'Unknown',
      cpu: parseFloat(cpuValue.slice(0, -1)) || 0, // Default to 0 if parsing fails
      memory: parseInt(memoryValue.slice(0, -2)) || 0 // Default to 0 if parsing fails
    };
  });

  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold mb-8">Kubernetes Metrics Dashboard</h1>
      
      <div className="mb-8">
        <h2 className="text-2xl font-semibold mb-4">Node Resources</h2>
        <div className="w-full h-[400px]">
          <BarChart data={nodeData} width={800} height={400}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="name" />
            <YAxis yAxisId="cpu" orientation="left" label={{ value: 'CPU (millicores)', angle: -90 }} />
            <YAxis yAxisId="memory" orientation="right" label={{ value: 'Memory (Mi)', angle: 90 }} />
            <Tooltip />
            <Legend />
            <Bar yAxisId="cpu" dataKey="cpu" fill="#8884d8" name="CPU" />
            <Bar yAxisId="memory" dataKey="memory" fill="#82ca9d" name="Memory" />
          </BarChart>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {metrics.pods.map(pod => (
          <div key={pod?.metadata?.name || Math.random()} className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
            <h3 className="text-lg font-semibold mb-2">{pod?.metadata?.name || 'Unknown Pod'}</h3>
            <div className="grid grid-cols-2 gap-2">
              <div>
                <span className="text-sm text-gray-500">CPU Usage</span>
                <p className="text-xl font-bold">{pod?.usage?.cpu || '0m'}</p>
              </div>
              <div>
                <span className="text-sm text-gray-500">Memory Usage</span>
                <p className="text-xl font-bold">{pod?.usage?.memory || '0Mi'}</p>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
