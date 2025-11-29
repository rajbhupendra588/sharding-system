import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import {
  Table,
  TableHead,
  TableHeader,
  TableBody,
  TableRow,
  TableCell,
} from './Table';

describe('Table Components', () => {
  describe('Table', () => {
    it('should render table wrapper', () => {
      render(
        <Table>
          <tbody>
            <tr>
              <td>Test</td>
            </tr>
          </tbody>
        </Table>
      );
      expect(screen.getByText('Test')).toBeInTheDocument();
    });

    it('should apply custom className', () => {
      render(
        <Table className="custom-table-class">
          <tbody>
            <tr>
              <td>Test</td>
            </tr>
          </tbody>
        </Table>
      );
      const table = screen.getByText('Test').closest('table');
      expect(table?.className).toContain('custom-table-class');
    });
  });

  describe('TableHead', () => {
    it('should render table head', () => {
      render(
        <Table>
          <TableHead>
            <TableHeader>Name</TableHeader>
          </TableHead>
          <TableBody>
            <TableRow>
              <TableCell>Test</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByText('Name')).toBeInTheDocument();
    });
  });

  describe('TableHeader', () => {
    it('should render table header', () => {
      render(
        <Table>
          <TableHead>
            <TableHeader>Column 1</TableHeader>
          </TableHead>
        </Table>
      );
      expect(screen.getByText('Column 1')).toBeInTheDocument();
    });

    it('should apply custom className', () => {
      render(
        <Table>
          <TableHead>
            <TableHeader className="custom-header-class">Header</TableHeader>
          </TableHead>
        </Table>
      );
      const header = screen.getByText('Header');
      expect(header.className).toContain('custom-header-class');
    });
  });

  describe('TableBody', () => {
    it('should render table body', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Row Data</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByText('Row Data')).toBeInTheDocument();
    });
  });

  describe('TableRow', () => {
    it('should render table row', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell Data</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByText('Cell Data')).toBeInTheDocument();
    });

    it('should handle onClick', () => {
      const handleClick = vi.fn();
      render(
        <Table>
          <TableBody>
            <TableRow onClick={handleClick}>
              <TableCell>Clickable Row</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      
      const row = screen.getByText('Clickable Row').closest('tr');
      if (row) {
        fireEvent.click(row);
        expect(handleClick).toHaveBeenCalledTimes(1);
      }
    });

    it('should apply cursor-pointer when onClick is provided', () => {
      render(
        <Table>
          <TableBody>
            <TableRow onClick={vi.fn()}>
              <TableCell>Clickable</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      
      const row = screen.getByText('Clickable').closest('tr');
      expect(row?.className).toContain('cursor-pointer');
    });

    it('should apply custom className', () => {
      render(
        <Table>
          <TableBody>
            <TableRow className="custom-row-class">
              <TableCell>Custom Row</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      
      const row = screen.getByText('Custom Row').closest('tr');
      expect(row?.className).toContain('custom-row-class');
    });
  });

  describe('TableCell', () => {
    it('should render table cell', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell>Cell Content</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByText('Cell Content')).toBeInTheDocument();
    });

    it('should apply colSpan attribute', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell colSpan={2}>Spanned Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      
      const cell = screen.getByText('Spanned Cell');
      expect(cell).toHaveAttribute('colSpan', '2');
    });

    it('should apply custom className', () => {
      render(
        <Table>
          <TableBody>
            <TableRow>
              <TableCell className="custom-cell-class">Custom Cell</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      
      const cell = screen.getByText('Custom Cell');
      expect(cell.className).toContain('custom-cell-class');
    });
  });

  describe('Complete Table Structure', () => {
    it('should render complete table structure', () => {
      render(
        <Table>
          <TableHead>
            <TableHeader>Name</TableHeader>
            <TableHeader>Status</TableHeader>
          </TableHead>
          <TableBody>
            <TableRow>
              <TableCell>Database 1</TableCell>
              <TableCell>Active</TableCell>
            </TableRow>
            <TableRow>
              <TableCell>Database 2</TableCell>
              <TableCell>Inactive</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );

      expect(screen.getByText('Name')).toBeInTheDocument();
      expect(screen.getByText('Status')).toBeInTheDocument();
      expect(screen.getByText('Database 1')).toBeInTheDocument();
      expect(screen.getByText('Database 2')).toBeInTheDocument();
      expect(screen.getByText('Active')).toBeInTheDocument();
      expect(screen.getByText('Inactive')).toBeInTheDocument();
    });
  });
});

