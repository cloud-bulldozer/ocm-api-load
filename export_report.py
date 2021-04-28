#!/usr/bin/python3

import plotly.io as pio
import pandas as pd
from docx import Document
from docx.shared import Inches
import os
import re
import sys
import datetime


def generate_graphs():
    graphs = {}
    # Walking through the folder to find all the result files
    # ignoring dirname since we don't need it in this implementation
    for root, _, files in os.walk("."):
        for f in files:
            regex = re.compile(r'[\w-]+_([\w-]+).(\w.+)')
            matches = regex.match(f)

            if 'reports' not in root and regex.match(f) and \
                    matches.group(2) == 'json':

                df = pd.read_json(root+'/'+f, lines=True)

                data = [{
                    'type': 'scatter',
                    'x': df['timestamp'],
                    'y': df['latency']/1000000,
                    'mode': 'markers',
                    'transforms': [{
                        'type': 'groupby',
                        'groups': df['code']
                    }]
                }]

                layout = {
                    'title': '<b>Latency per Request: {}</b>'.format(
                        matches.group(1)),
                    'xaxis': {'title': 'Time',
                              'showgrid': 'true',
                              'ticklabelmode': "period"},
                    'yaxis': {'title': 'Milliseconds (log)',
                              'type': 'log'},
                }

                fig_dict = {'data': data, 'layout': layout}

                pio.write_image(fig_dict,
                                root+'/'+matches.group(1)+".png",
                                engine="kaleido",
                                width=1600,
                                height=900,
                                validate=False)
                graphs[matches.group(1)] = root+'/'+matches.group(1)+".png"
    return graphs


def show_graphs(file):
    regex = re.compile(r'[\w-]+_([\w-]+).(\w.+)')
    matches = regex.match(file)

    if regex.match(file) and matches.group(2) == 'json':
        df = pd.read_json(file, lines=True)

        data = [{
            'type': 'scatter',
            'x': df['timestamp'],
            'y': df['latency']/1000000,
            'mode': 'markers',
            'transforms': [
                {'type': 'groupby',
                 'groups': df['code']}]
                }]

        layout = {
            'title': '<b>Latency per Request: {}</b>'.format(matches.group(1)),
            'xaxis': {'title': 'Time',
                      'showgrid': 'true',
                      'ticklabelmode': "period"},
            'yaxis': {'title': 'Milliseconds (log)', 'type': 'log'},
        }

        fig_dict = {'data': data, 'layout': layout}

        pio.show(fig_dict,
                 engine="kaleido",
                 width=1600,
                 height=900,
                 validate=False)


def read_reports():
    reports = {}
    # Walking through the folder to find all the report files
    # ignoring dirname since we don't need it in this implementation
    for root, _, files in os.walk("."):
        for f in files:
            regex = re.compile(
                r'[\w-]+_([\w-]+)-report-\d{4}-\d{2}-\d{2}.(\w.+)')
            matches = regex.match(f)

            if 'reports' in root and regex.match(f) and \
                    matches.group(2) == 'json':
                df = pd.read_json(root+'/'+f, lines=True)

                lat = df['latencies'][0]
                reports[matches.group(1)] = {
                    'requests':     int(df['requests']),
                    'rate':         float(df['rate']),
                    'duration':     int(df['duration']),
                    'min':          int(lat['min']),
                    'mean':         int(lat['mean']),
                    'max':          int(lat['max']),
                    'success':      float(df['success']),
                    'status_codes': df['status_codes'][0],
                    'errors':       df['errors'][0],
                }
    return reports


def write_docx(reports, graphs):
    date = datetime.datetime.utcnow()
    document = Document()

    document.add_heading('OCM Performance Test', 0)

    document.add_heading('Test # ', level=1)
    document.add_paragraph('Date: {}'.format(date.strftime("%Y-%m-%d")))

    document.add_heading('Description', level=2)
    document.add_paragraph('The purpose of this test is ...')

    document.add_heading('Notes', level=3)

    document.add_heading('Endpoints', level=2)

    table = document.add_table(rows=1, cols=3)
    hdr_cells = table.rows[0].cells
    hdr_cells[0].text = 'Enpoint'
    hdr_cells[1].text = 'Rate'
    hdr_cells[2].text = 'Notes'
    for r in reports:
        row_cells = table.add_row().cells
        row_cells[0].text = r
        row_cells[1].text = '{}/s for {} minutes'.format(
            reports[r]['rate'], reports[r]['duration'])
        row_cells[2].text = ''

    document.add_heading('Per endpoint data', level=2)
    for r in reports:
        document.add_heading('{}'.format(r), level=3)
        document.add_picture(graphs[r], width=Inches(16.6), height=Inches(9.4))
        p = document.add_paragraph(
            'Requests\t\tTotal: {}\t\tRate: {}\n'.format(
                reports[r]['requests'], reports[r]['rate']))
        p.add_run('Duration\t\t{}\n'.format(reports[r]['duration']))
        p.add_run('Latencies\n')

        document.add_paragraph('Min: {}ms'.format(
            reports[r]['min']), style='List Bullet')
        document.add_paragraph('Mean: {}ms'.format(
            reports[r]['mean']), style='List Bullet')
        document.add_paragraph('Max: {}ms'.format(
            reports[r]['max']), style='List Bullet')

        p2 = document.add_paragraph('Success\t\t{}%\n'.format(
            reports[r]['success']))
        p2.add_run('Status Codes\t\t\n{}\n'.format(reports[r]['status_codes']))
        p2.add_run('Error Set\t\t\n{}\n'.format(reports[r]['errors']))
        p2.add_run('Notes').bold = True
        p2.add_run('\n')
        document.add_page_break()
    document.add_heading('Conclusion', level=2)
    document.add_paragraph('Make sure....', style='List Bullet')
    document.add_page_break()
    document.add_heading('Overall Screenshots', level=2)
    document.save('report-{}.docx'.format(date.strftime("%Y-%m-%d")))


def main():
    args = sys.argv[1:]
    if len(args) > 1 and args[0] == 'graph':
        show_graphs(args[1])
        exit()

    graphs = generate_graphs()
    reports = read_reports()
    write_docx(reports, graphs)


if __name__ == "__main__":
    main()
