import json
import os
import sys

SPEC_REPORTS_KEY = 'SpecReports'
REPORT_ENTRIES_KEY = 'ReportEntries'
REPORT_ENTRY_VALUE_KEY = 'Value'
REPORT_ENTRY_VALUE_JSON_KEY = 'AsJSON'

VALUES_KEY = 'Values'
ANNOTATIONS_KEY = 'Annotations'
MEASUREMENTS_KEY = 'Measurements'
ANNOTATED_VALUES_KEY = 'AnnotatedValues'

def extract_json_objects(report_file):
    json_objects = []
    benchmark_reports = {}

    with open(report_file, 'r') as f:
        benchmark_reports = json.load(f)

    for benchmark_report in benchmark_reports:
        spec_reports = benchmark_report.get(SPEC_REPORTS_KEY, [])

        for spec_report in spec_reports:
            report_entries = spec_report.get(REPORT_ENTRIES_KEY, [])

            for entry in report_entries:
                value = entry.get(REPORT_ENTRY_VALUE_KEY, {})
                json_str = value.get(REPORT_ENTRY_VALUE_JSON_KEY, '{}')

                json_objects.append(json.loads(json_str))

    return json_objects

def add_annotation_value_mapping(json_objects):
    mapped_objects = []

    for json_object in json_objects:
        mapped_measurements = []
        measurements = json_object.get(MEASUREMENTS_KEY, [])

        for measurement in measurements:
            mapping = {}

            values = measurement.get(VALUES_KEY, [])
            annotations = measurement.get(ANNOTATIONS_KEY, [])

            # Annotations and Values are parallel lists
            for i in range(len(annotations)):
                key = annotations[i]
                
                if mapping.get(key) is None:
                    mapping[key] = []
                
                mapping[key].append(values[i])

            mapped_measurement = measurement
            mapped_measurement[ANNOTATED_VALUES_KEY] = mapping

            mapped_measurements.append(mapped_measurement)

        mapped_object = {}
        mapped_object[MEASUREMENTS_KEY] = mapped_measurements

        mapped_objects.append(mapped_object)

    return mapped_objects

def main():
    if len(sys.argv) <= 1:
        print("error: no ginkgo results directory given")
        sys.exit(-1)

    results_directory = sys.argv[1]

    benchmark_file = os.path.join(results_directory, 'report.json')
    output_file = os.path.join(results_directory, 'measurements.json')

    if not os.path.exists(benchmark_file):
        print ("no report file found.")
        return
    
    objects = extract_json_objects(benchmark_file)
    mapped_objects = add_annotation_value_mapping(objects)

    with open(output_file, 'w') as f:
        json.dump(mapped_objects, f, indent=4)

main()