#!/usr/bin/env python3
import json
import requests
import time
import sys
import matplotlib.pyplot as plt
import numpy as np

class MapReduceOrchestrator:
    def __init__(self, splitter_url, mapper_urls, reducer_url):
        self.splitter_url = splitter_url
        self.mapper_urls = mapper_urls
        self.reducer_url = reducer_url
        self.timings = {
            'split': 0,
            'map': [],
            'reduce': 0,
            'total': 0
        }
    
    def run_word_count(self, s3_url, chunks=3):
        """Run the complete MapReduce word count pipeline"""
        start_time = time.time()
        
        print(f"Starting MapReduce word count on: {s3_url}")
        print(f"Number of chunks: {chunks}")
        print("-" * 50)
        
        # Step 1: Split the file
        print("1. Splitting file into chunks...")
        split_response = self._split_file(s3_url, chunks)
        chunk_urls = split_response['chunk_urls']
        self.timings['split'] = split_response['processing_time_ms']
        print(f"   Split completed in {self.timings['split']:.2f}ms")
        print(f"   Created {len(chunk_urls)} chunks")
        
        # Step 2: Map each chunk
        print("\n2. Mapping chunks in parallel...")
        map_results = self._map_chunks(chunk_urls)
        print(f"   Mapping completed")
        print(f"   Average map time: {np.mean(self.timings['map']):.2f}ms")
        
        # Step 3: Reduce results
        print("\n3. Reducing results...")
        final_result = self._reduce_results(map_results)
        self.timings['reduce'] = final_result['processing_time_ms']
        print(f"   Reduce completed in {self.timings['reduce']:.2f}ms")
        
        # Calculate total time
        self.timings['total'] = (time.time() - start_time) * 1000
        
        # Print results
        print("\n" + "=" * 50)
        print("RESULTS:")
        print(f"Total unique words: {final_result['unique_words']}")
        print(f"Total word count: {final_result['total_words']}")
        print(f"Final results saved to: {final_result['final_url']}")
        print("\nTop 10 words:")
        for i, word_info in enumerate(final_result['top_10_words'], 1):
            print(f"  {i:2d}. {word_info['word']:15s} - {word_info['count']:,} occurrences")
        
        print("\n" + "=" * 50)
        print("PERFORMANCE METRICS:")
        print(f"Split time:    {self.timings['split']:8.2f}ms")
        print(f"Map time (avg): {np.mean(self.timings['map']):8.2f}ms")
        print(f"Map time (max): {max(self.timings['map']):8.2f}ms")
        print(f"Reduce time:   {self.timings['reduce']:8.2f}ms")
        print(f"Total time:    {self.timings['total']:8.2f}ms")
        
        return final_result, self.timings
    
    def _split_file(self, s3_url, chunks):
        """Call the splitter service"""
        response = requests.post(
            f"{self.splitter_url}/split",
            json={"s3_url": s3_url, "chunks": chunks}
        )
        response.raise_for_status()
        return response.json()
    
    def _map_chunks(self, chunk_urls):
        """Call mapper services in parallel"""
        import concurrent.futures
        
        def map_single_chunk(mapper_url, chunk_url):
            response = requests.post(
                f"{mapper_url}/map",
                json={"chunk_url": chunk_url}
            )
            response.raise_for_status()
            return response.json()
        
        # Map chunks to mappers (round-robin)
        with concurrent.futures.ThreadPoolExecutor(max_workers=len(self.mapper_urls)) as executor:
            futures = []
            for i, chunk_url in enumerate(chunk_urls):
                mapper_url = self.mapper_urls[i % len(self.mapper_urls)]
                future = executor.submit(map_single_chunk, mapper_url, chunk_url)
                futures.append(future)
            
            results = []
            for future in concurrent.futures.as_completed(futures):
                result = future.result()
                results.append(result['result_url'])
                self.timings['map'].append(result['processing_time_ms'])
        
        return results
    
    def _reduce_results(self, result_urls):
        """Call the reducer service"""
        response = requests.post(
            f"{self.reducer_url}/reduce",
            json={"result_urls": result_urls}
        )
        response.raise_for_status()
        return response.json()
    
    def run_performance_experiments(self, s3_url, chunk_sizes=[2, 3, 4, 5, 6]):
        """Run experiments with different chunk sizes"""
        results = []
        
        for chunks in chunk_sizes:
            print(f"\n{'='*60}")
            print(f"Running experiment with {chunks} chunks")
            print(f"{'='*60}")
            
            try:
                _, timings = self.run_word_count(s3_url, chunks)
                results.append({
                    'chunks': chunks,
                    'split_time': timings['split'],
                    'map_time_avg': np.mean(timings['map']),
                    'map_time_max': max(timings['map']),
                    'reduce_time': timings['reduce'],
                    'total_time': timings['total']
                })
            except Exception as e:
                print(f"Error with {chunks} chunks: {e}")
        
        return results
    
    def plot_results(self, results):
        """Create performance visualization"""
        if not results:
            print("No results to plot")
            return
        
        chunks = [r['chunks'] for r in results]
        split_times = [r['split_time'] for r in results]
        map_times = [r['map_time_max'] for r in results]
        reduce_times = [r['reduce_time'] for r in results]
        total_times = [r['total_time'] for r in results]
        
        fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(12, 10))
        
        # Plot 1: Component times
        ax1.plot(chunks, split_times, 'o-', label='Split')
        ax1.plot(chunks, map_times, 's-', label='Map (max)')
        ax1.plot(chunks, reduce_times, '^-', label='Reduce')
        ax1.set_xlabel('Number of Chunks')
        ax1.set_ylabel('Time (ms)')
        ax1.set_title('Component Processing Times')
        ax1.legend()
        ax1.grid(True)
        
        # Plot 2: Total time
        ax2.plot(chunks, total_times, 'o-', color='red', linewidth=2)
        ax2.set_xlabel('Number of Chunks')
        ax2.set_ylabel('Total Time (ms)')
        ax2.set_title('Total Processing Time vs Chunks')
        ax2.grid(True)
        
        # Plot 3: Stacked bar chart
        width = 0.6
        x = np.arange(len(chunks))
        ax3.bar(x, split_times, width, label='Split')
        ax3.bar(x, map_times, width, bottom=split_times, label='Map')
        ax3.bar(x, reduce_times, width, 
                bottom=np.array(split_times) + np.array(map_times), label='Reduce')
        ax3.set_xlabel('Number of Chunks')
        ax3.set_ylabel('Time (ms)')
        ax3.set_title('Time Distribution by Component')
        ax3.set_xticks(x)
        ax3.set_xticklabels(chunks)
        ax3.legend()
        
        # Plot 4: Speedup analysis
        baseline = total_times[0] if total_times else 1
        speedup = [baseline / t for t in total_times]
        ax4.plot(chunks, speedup, 'o-', color='green', linewidth=2)
        ax4.axhline(y=1, color='r', linestyle='--', alpha=0.5)
        ax4.set_xlabel('Number of Chunks')
        ax4.set_ylabel('Speedup')
        ax4.set_title('Speedup vs Baseline (2 chunks)')
        ax4.grid(True)
        
        plt.tight_layout()
        plt.savefig('mapreduce_performance.png', dpi=150)
        plt.show()

def main():
    # Configure your ECS task public IPs here
    SPLITTER_URL = "http://18.237.143.160:8080"
    MAPPER_URLS = [
        "http://35.90.230.22:8080",
        "http://44.244.74.12:8080",
        "http://35.91.111.55:8080"
    ]
    REDUCER_URL = "http://34.212.38.161:8080"
    
    # S3 URL of your text file
    S3_URL = "s3://mapreduce-wordcount-730335606003/input/hamlet.txt"
    
    orchestrator = MapReduceOrchestrator(SPLITTER_URL, MAPPER_URLS, REDUCER_URL)
    
    # Run single test
    if len(sys.argv) > 1 and sys.argv[1] == "test":
        orchestrator.run_word_count(S3_URL, chunks=3)
    else:
        # Run performance experiments
        results = orchestrator.run_performance_experiments(S3_URL, [2, 3, 4, 5, 6])
        orchestrator.plot_results(results)
        
        # Save results
        with open('performance_results.json', 'w') as f:
            json.dump(results, f, indent=2)
        print("\nResults saved to performance_results.json")

if __name__ == "__main__":
    main()