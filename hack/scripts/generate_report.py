import json
import matplotlib.pyplot as plt
from jinja2 import Template
import os
import argparse
import yaml

# Parse command-line arguments
parser = argparse.ArgumentParser(description='Generate benchmark report from measurements.json')
parser.add_argument('dir_paths', type=str, nargs='+', help='Paths to the directories containing the measurements.json files')
args = parser.parse_args()

# Function to load benchmark description from benchmark.yaml
# Function to load benchmark description from benchmark.yaml
def load_benchmark_description(dir_path):
    yaml_path = os.path.join(dir_path, 'benchmark.yaml')
    with open(yaml_path) as f:
        benchmark_data = yaml.safe_load(f)
    
    ingestion_description = benchmark_data.get('scenarios', {}).get('ingestionPath', {}).get('description')
    if ingestion_description:
        return ingestion_description
    
    query_description = benchmark_data.get('scenarios', {}).get('queryPath', {}).get('description', 'Unknown Benchmark')
    return query_description
    
# Function to plot a measurement and save as image
def plot_measurement(measurements, output_dir, plot_index):
    plt.figure(figsize=(10, 6))
    for measurement, description in measurements:
        name = measurement['Name']
        values = measurement['Values']
        units = measurement['Units']
        annotations = measurement.get('Annotations', [])
        
        # Generate time values for x-axis starting from 3 minutes
        time_values = [(i + 1) * 3 for i in range(len(values))]
        
        plt.plot(time_values, values, marker='o', label=description)
    
    plt.title(f'{name}')
    plt.xlabel('Time (minutes)')
    plt.ylabel(f'{units}')
    plt.legend()
    plt.grid(True)
    
    plot_filename = os.path.join(output_dir, f'plot_{plot_index}.png')
    plt.savefig(plot_filename)
    plt.close()
    
    return f'./plots/plot_{plot_index}.png', f'{name}'

# Collect all measurements from the provided directories
all_measurements = {}
for dir_path in args.dir_paths:
    json_path = os.path.join(dir_path, 'measurements.json')
    with open(json_path) as f:
        data = json.load(f)
    
    benchmark_description = load_benchmark_description(dir_path)
    measurements = data[0]['Measurements']
    
    for measurement in measurements:
        name = measurement['Name']
        if name not in all_measurements:
            all_measurements[name] = []
        all_measurements[name].append((measurement, benchmark_description))

# Determine the parent directory for the README and plots
parent_dir = os.path.commonpath(args.dir_paths)
output_dir = os.path.join(parent_dir, 'plots')
os.makedirs(output_dir, exist_ok=True)

# Plot all measurements and save images
plot_files = []
for plot_index, (name, measurements) in enumerate(all_measurements.items()):
    plot_file, plot_title = plot_measurement(measurements, output_dir, plot_index)
    plot_files.append((plot_title, plot_file))

# Load README template
template_path = 'reports/README.template'
with open(template_path) as f:
    template_content = f.read()

# Render README with plots
template = Template(template_content)
rendered_readme = template.render(plots=plot_files)

# Save rendered README
readme_path = os.path.join(parent_dir, 'README.md')
with open(readme_path, 'w') as f:
    f.write(rendered_readme)

print(f"Plots and README.md generated successfully in {parent_dir}.")